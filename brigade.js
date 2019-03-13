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
// TODO: we should determine which check run is being requested and *only* run this
events.on("check_run:rerequested", runSuite);

events.on("exec", (e, p) => {
  Group.runAll([
    build(e, p),
    xbuild(e, p),
    test(e, p),
    testIntegration(e, p)
  ]);
});

// Although a GH App will trigger 'check_suite:requested' on a push to master event,
// it will not for a tag push, hence the need for this handler
events.on("push", (e, p) => {
  if (e.revision.ref.includes("refs/heads/master") || e.revision.ref.startsWith("refs/tags/")) {
    publish(e, p).run();
  }
});

events.on("publish", (e, p) => {
  publish(e, p).run();
});

// **********************************************
// Actions
// **********************************************

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

  // Setup kubeconfig, fetch duffle, docker login, run tests
  goTest.tasks.push(
    "mkdir -p ${HOME}/.kube",
    'echo "${kubeconfig}" > ${HOME}/.kube/config',
    "apt-get update && apt-get install -y sudo",
    "curl -fsSL https://raw.githubusercontent.com/fishworks/gofish/master/scripts/install.sh | bash",
    "gofish init && gofish install duffle",
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
    "make xbuild-all publish"
  )

  return goPublish;
}

// Here we add GitHub Check Runs, which will run in parallel and report their results independently to GitHub
function runSuite(e, p) {
  checkRun(e, p, build, "Build").catch(e  => {console.error(e.toString())});
  checkRun(e, p, xbuild, "Cross-Platform Build").catch(e  => {console.error(e.toString())});
  checkRun(e, p, test, "Test").catch(e  => {console.error(e.toString())});
  checkRun(e, p, testIntegration, "Integration Test").catch(e  => {console.error(e.toString())});
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
    this.detailsURL = `https://azure.github.io/kashti/builds/${ e.buildID }`;
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