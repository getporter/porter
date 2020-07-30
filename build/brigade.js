const { events, Job, Group } = require("brigadier");
const { KindJob } = require("@brigadecore/brigade-utils");

// **********************************************
// Globals
// **********************************************

const projectName = "porter";

// **********************************************
// Event Handlers
// **********************************************

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", runCheck);
events.on("issue_comment:created", handleIssueComment);
events.on("issue_comment:edited", handleIssueComment);

events.on("exec", (e, p) => {
  return Group.runAll([
    build(e, p),
    xbuild(e, p),
    testUnit(e, p),
    testIntegration(e, p),
    testCLI(e, p),
    validate(e, p)
  ]);
});

events.on("test-integration", (e, p) => {
  return Group.runAll([
    testIntegration(e, p),
    testCLI(e, p)
  ]);
})

// Although a GH App will trigger 'check_suite:requested' on a push to main branch event,
// it will not for a tag push, hence the need for this handler
events.on("push", (e, p) => {
  if (e.revision.ref.startsWith("refs/tags/")) {
    return Group.runEach([
      testUnit(e, p),
      testIntegration(e, p),
      testCLI(e, p),
      publish(e, p),
      publishExamples(e, p)
    ])
  }
});

events.on("publish", (e, p) => {
  return Group.runEach([
    testUnit(e, p),
    testIntegration(e, p),
    testCLI(e, p),
    publish(e, p),
    publishExamples(e, p)
  ])
});

events.on("publish-examples", (e, p) => {
  return publishExamples(e, p).run();
})

// **********************************************
// Actions
// **********************************************

// Important: each Job name below must only consist of lowercase
// alphanumeric characters and hyphens, per K8s resource name restrictions
function build(e, p) {
  var goBuild = new GoJob(`${projectName}-build`);

  goBuild.tasks.push(
    "make build"
  );

  return goBuild;
}

function validate(e, p) {
  var validator = new GoJob(`${projectName}-validate`);
  // Enable Docker-in-Docker (needed for building bundles)
  validator.enableDind();

  validator.tasks.push(
    "apk add --update npm",
    "npm install -g ajv-cli",
    "make build install",
    "make build-bundle validate-bundle"
  );

  return validator;
}

function xbuild(e, p) {
  var goBuild = new GoJob(`${projectName}-xbuild`);

  goBuild.tasks.push(
    "make xbuild-all"
  );

  return goBuild;
}

function testUnit(e, p) {
  var goTest = new GoJob(`${projectName}-testunit`);

  goTest.tasks.push(
    "make test-unit"
  );

  return goTest;
}

// TODO: we could refactor so that this job shares a mount with the build job above,
// to remove the need of re-building before running test-cli
function testIntegration(e, p) {
  var testInt = new KindJob(`${projectName}-testintegration`, "vdice/go-dind:kind-v0.7.0");

  testInt.tasks.push(
    "mkdir -p /go/bin",
    "cd /src",
    "trap 'make -f Makefile.kind delete-kind-cluster' EXIT",
    `make -f Makefile.kind create-kind-cluster`,
    "make test-integration"
  );

  return testInt;
}

function testCLI(e, p) {
  var testCLI = new GoJob(`${projectName}-testcli`);
  // Enable Docker-in-Docker
  testCLI.enableDind();

  testCLI.tasks.push(
    "make test-cli"
  );

  return testCLI;
}

function publishExamples(e, p) {
  var examplePublisher = new GoJob(`${projectName}-publish-examples`);
  // Enable Docker-in-Docker
  examplePublisher.enableDind();

  examplePublisher.tasks.push(
    "apk add --update npm",
    "npm install -g ajv-cli",
    // first, build and install porter
    `make build install`,
    // login to the registry we'll be pushing to
    `docker login ${p.secrets.dockerhubRegistry} \
    -u ${p.secrets.dockerhubUsername} \
    -p ${p.secrets.dockerhubPassword}`,
    // now build, validate and publish bundles
    `REGISTRY=${p.secrets.dockerhubOrg} make build-bundle validate-bundle publish-bundle`
  );

  return examplePublisher;
}

function publish(e, p) {
  var goPublish = new GoJob(`${projectName}-publish`);

  // TODO: we could/should refactor so that this job shares a mount with the xbuild job above,
  // to remove the need of re-xbuilding before publishing

  goPublish.env.AZURE_STORAGE_CONNECTION_STRING = p.secrets.azureStorageConnectionString;
  goPublish.tasks.push(
    // Fetch az cli needed for publishing
    "curl -sLO https://github.com/carolynvs/az-cli/releases/download/v0.3.2/az-linux-amd64 && \
      chmod +x az-linux-amd64 && \
      mv az-linux-amd64 /usr/local/bin/az",
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
  "build": { runFunc: build, description: "Build" },
  "validate": { runFunc: validate, description: "Validate" },
  "crossplatformbuild": { runFunc: xbuild, description: "Cross-Platform Build" },
  "unittest": { runFunc: testUnit, description: "Unit Test" },
  "integrationtest": { runFunc: testIntegration, description: "Integration Test" },
  "clitest": { runFunc: testCLI, description: "CLI Test" }
};

// runCheck can be invoked to (re-)run an individual GitHub Check Run, as opposed to a full suite
function runCheck(e, p) {
  payload = JSON.parse(e.payload);

  name = payload.body.check_run.name;
  check = checks[name];

  if (typeof check !== 'undefined') {
    return checkRun(e, p, check.runFunc, check.description);
  } else {
    throw new Error(`No check found with name: ${name}`);
  }
}

// Here we add GitHub Check Runs, which will run in parallel and report their results independently to GitHub
function runSuite(e, p) {
  var checkRuns = new Array();

  // Construct Check Run Suite depending on branch
  if (e.revision.ref == "main" ) {
    checkRuns = [
      checkRun(e, p, testUnit, "Unit Test"),
      checkRun(e, p, testIntegration, "Integration Test"),
      checkRun(e, p, testCLI, "CLI Test"),
      checkRun(e, p, publish, "Publish"),
      checkRun(e, p, publishExamples, "Publish Example Bundles")
    ]
  } else {
    for (check of Object.values(checks)) {
      checkRuns.push(checkRun(e, p, check.runFunc, check.description));
    }
  }

  // Now run the Check Run Suite
  //
  // Important: To prevent Promise.all() from failing fast, we catch and
  // return all errors. This ensures Promise.all() always resolves. We then
  // iterate over all resolved values looking for errors. If we find one, we
  // throw it so the whole build will fail.
  //
  // Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all#Promise.all_fail-fast_behaviour
  return Promise.all(checkRuns)
    .then((values) => {
      values.forEach((value) => {
        if (value instanceof Error) throw value;
      });
    });
}

// handleIssueComment handles an issue_comment event, parsing the comment
// text and determining whether or not to trigger a corresponding action
function handleIssueComment(e, p) {
  payload = JSON.parse(e.payload);

  // Extract the comment body and trim whitespace
  comment = payload.body.comment.body.trim();

  // Here we determine if a comment should provoke an action
  switch(comment) {
    case "/brig run":
      runSuite(e, p);
      break;
    default:
      console.log(`No applicable action found for comment: ${comment}`);
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
    const localPath = gopath + `/src/get.porter.sh/${projectName}`;

    this.image = "quay.io/vdice/go-dind:v0.1.2";

    this.tasks = [
      // Need to move the source into GOPATH so vendor/ works as desired.
      "mkdir -p " + localPath,
      "mv /src/* " + localPath,
      "mv /src/.git " + localPath,
      "cd " + localPath
    ];

    // Set default job timeout to 1800000 milliseconds / 30 minutes
    this.timeout = 1800000
  }

  // enabledDind enables Docker-in-Docker for this job,
  // setting privileged to true and adding daemon setup to tasks list
  enableDind() {
    this.privileged = true;
    this.tasks.push(
      "dockerd-entrypoint.sh &",
      "sleep 20");
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
    var j = new Job(`${ this.name }-${ this.count }`, "brigadecore/brigade-github-check-run:v0.1.0");
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
    let res = await job.run();
    const logs = await job.logs();

    note.conclusion = conclusion;
    note.summary = `Task "${ job.name }" passed`;
    note.text = `Task Complete: ${conclusion}`;
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
