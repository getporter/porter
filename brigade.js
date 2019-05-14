const { events, Job, Group } = require("brigadier");

// **********************************************
// Globals
// **********************************************

const projectName = "porter";
const projectOrg = "deislabs";

// **********************************************
// Event Handlers
// **********************************************

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", runCheck);

events.on("exec", (e, p) => {
  Group.runAll([
    verify(e, p),
    build(e, p),
    xbuild(e, p),
    test(e, p),
    testIntegration(e, p)
  ]);
});

// Although a GH App will trigger 'check_suite:requested' on a push to master event,
// it will not for a tag push, hence the need for this handler
events.on("push", (e, p) => {
  if (e.revision.ref.startsWith("refs/tags/")) {
    Group.runEach([
      test(e, p),
      testIntegration(e, p),
      publish(e, p),
    ])
  }
});

events.on("publish", (e, p) => {
  Group.runEach([
    test(e, p),
    testIntegration(e, p),
    publish(e, p),
  ])
});

// **********************************************
// Actions
// **********************************************

function verify(e, p) {
  var goBuild = new GoJob(`${projectName}-verify`);

  goBuild.tasks.push(
      "make verify"
  );

  return goBuild;
}

function build(e, p) {
  var goBuild = new GoJob(`${projectName}-build`);

  goBuild.tasks.push(
    "make build"
  );

  return goBuild;
}

function xbuild(e, p) {
  var goBuild = new GoJob(`${projectName}-xbuild`);

  goBuild.tasks.push(
    "make xbuild-all"
  );

  return goBuild;
}

function test(e, p) {
  var goTest = new GoJob(`${projectName}-test`);

  goTest.tasks.push(
    "make test-unit"
  );

  return goTest;
}

// TODO: we could refactor so that this job shares a mount with the build job above,
// to remove the need of re-building before running test-cli
function testIntegration(e, p) {
  var goTest = new GoJob(`${projectName}-integrationtest`);
  // Enable docker so that the daemon can be used for duffle commands invoked by test-cli
  goTest.docker.enabled = true;

  goTest.env.kubeconfig = {
    secretKeyRef: {
      name: "porter-kubeconfig",
      key: "kubeconfig"
    }
  };

  // Setup kubeconfig, docker login, run tests
  goTest.tasks.push(
    "mkdir -p ${HOME}/.kube",
    'echo "${kubeconfig}" > ${HOME}/.kube/config',
    `docker login ${p.secrets.dockerhubRegistry} \
      -u ${p.secrets.dockerhubUsername} \
      -p ${p.secrets.dockerhubPassword}`,
    `REGISTRY=${p.secrets.dockerhubOrg} make test-cli`
  );

  return goTest;
}

function publish(e, p) {
  var goPublish = new GoJob(`${projectName}-publish`);

  // TODO: we could/should refactor so that this job shares a mount with the xbuild job above,
  // to remove the need of re-xbuilding before publishing

  goPublish.env.AZURE_STORAGE_CONNECTION_STRING = p.secrets.azureStorageConnectionString;
  goPublish.tasks.push(
    "make build xbuild-all publish"
  )

  return goPublish;
}

// These represent the standard checks that run for PRs (see runSuite below)
// 
// They are encapsulated in an object such that individual checks may be run
// using data supplied by a corresponding GitHub webhook, say, on a
// check_run:rerequested event (see runCheck below)
checks = {
  "verify": { runFunc: verify, description: "Verify" },
  "build": { runFunc: build, description: "Build" },
  "crossplatformbuild": { runFunc: xbuild, description: "Cross-Platform Build" },
  "test": { runFunc: test, description: "Test" },
  "integrationtest": { runFunc: testIntegration, description: "Integration Test" }
};

// runCheck can be invoked to (re-)run an individual GitHub Check Run, as opposed to a full suite
function runCheck(e, p) {
  payload = JSON.parse(e.payload);

  name = payload.body.check_run.name;
  check = checks[name];

  if (typeof check !== 'undefined') {
    checkRun(e, p, check.runFunc, check.description)
      .catch(e => {console.error(e.toString())});
  } else {
    err = new Error(`No check found with name: ${name}`);
    // TODO: remove this console.error statement once Brigade logs err.message when thrown
    // https://github.com/brigadecore/brigade-github-app/pull/43
    console.error(err.message);
    throw err;
  }
}

// Here we add GitHub Check Runs, which will run in parallel and report their results independently to GitHub
function runSuite(e, p) {
  if (e.revision.ref == "master" ) {
    Group.runEach([
      checkRun(e, p, test, "Test"),
      checkRun(e, p, testIntegration, "Integration Test"),
      checkRun(e, p, publish, "Publish")
    ])
  } else {
    for (check of Object.values(checks)) {
        checkRun(e, p, check.runFunc, check.description)
          .catch(e => {console.error(e.toString())});
    }
  }
}

// **********************************************
// Classes/Helpers
// **********************************************

// GoJob is a Job with Golang-related prerequisites set up
class GoJob extends Job {
  constructor (name) {
    super(name);

    const gopath = "/go";
    const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`;

    // Here using the large-but-useful deis/go-dev image as we have a need for deps
    // already pre-installed in this image, e.g. helm, az, docker, etc.
    // TODO: replace with lighter-weight image (Carolyn)
    this.image = "deis/go-dev";
    this.env = {
      "GOPATH": gopath
    };
    this.tasks = [
      // Need to move the source into GOPATH so vendor/ works as desired.
      "mkdir -p " + localPath,
      "mv /src/* " + localPath,
      "mv /src/.git " + localPath,
      "cd " + localPath
    ];
    this.streamLogs = true;
  }
}

// checkRun is a GitHub Check Run that is ran as part of a Checks Suite,
// running the provided runFunc corresponding to the provided description 
function checkRun(e, p, runFunc, description) {
  console.log(`Check requested: ${description}`);

  // Create Notification object (which is just a Job to update GH using the Checks API)
  note = new Notification(description.toLowerCase().replace(/[^a-z]/g, ''), e, p);
  note.conclusion = "";
  note.title = `Run ${description}`;
  note.summary = `Running ${description} for ${e.revision.commit}`;
  note.text = `Ensuring ${description} complete(s) successfully`

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(runFunc(e, p), note);
}

// A GitHub Check Suite notification
class Notification {
  constructor(name, e, p) {
    this.proj = p;
    this.payload = e.payload;
    this.name = name;
    this.externalID = e.buildID;
    this.detailsURL = `https://brigadecore.github.io/kashti/builds/${ e.buildID }`;
    this.title = "running check";
    this.text = "";
    this.summary = "";

    // count allows us to send the notification multiple times, with a distinct pod name
    // each time.
    this.count = 0;

    // One of: "success", "failure", "neutral", "cancelled", or "timed_out".
    this.conclusion = "neutral";
  }

  // Send a new notification, and return a Promise<result>.
  run() {
    this.count++
    var j = new Job(`${ this.name }-${ this.count }`, "deis/brigade-github-check-run:latest");
    j.env = {
      CHECK_CONCLUSION: this.conclusion,
      CHECK_NAME: this.name,
      CHECK_TITLE: this.title,
      CHECK_PAYLOAD: this.payload,
      CHECK_SUMMARY: this.summary,
      CHECK_TEXT: this.text,
      CHECK_DETAILS_URL: this.detailsURL,
      CHECK_EXTERNAL_ID: this.externalID
    }
    return j.run();
  }
}

// notificationWrap is a helper to wrap a job execution between two notifications.
async function notificationWrap(job, note, conclusion) {
  if (conclusion == null) {
    conclusion = "success"
  }
  await note.run();
  try {
    let res = await job.run()
    const logs = await job.logs();

    note.conclusion = conclusion;
    note.summary = `Task "${ job.name }" passed`;
    note.text = note.text = `Task Complete: ${conclusion}`;
    return await note.run();
  } catch (e) {
    const logs = await job.logs();
    note.conclusion = "failure";
    note.summary = `Task "${ job.name }" failed for ${ e.buildID }`;
    note.text = "Task failed with error: " + e.toString();
    try {
      return await note.run();
    } catch (e2) {
      console.error("failed to send notification: " + e2.toString());
      console.error("original error: " + e.toString());
      return e2;
    }
  }
}
