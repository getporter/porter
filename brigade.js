const { events, Job, Group } = require("brigadier");

// TODO: place globals, helpers, etc. in different js files?
// e.g.:
// const { GoJob, Notification } = require("helpers")

// Globals

const projectName = "porter";
const projectOrg = "deislabs";

class GoJob extends Job {
  constructor (name) {
    super(name);

    const gopath = "/go";
    const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`;

    this.image = "golang:1.11";
    this.env = {
      "GOPATH": gopath
    };
    this.tasks = [
      // Need to move the source into GOPATH so vendor/ works as desired.
      "mkdir -p " + localPath,
      "mv /src/* " + localPath,
      "mv /src/.git " + localPath,
      "cd " + localPath,
      "make get-deps"
    ];
    this.streamLogs = true;
  }
}

// Event Handlers

events.on("exec", (e, p) => {
  Group.runAll([
    // build(e, p),
    // xbuild(e, p),
    // test(e, p),
    testIntegration(e, p)
  ]);
})

events.on("check_suite:requested", runSuite)
events.on("check_suite:rerequested", runSuite)
events.on("check_run:rerequested", runSuite)

// Although a GH App will trigger 'check_suite:requested' on a push to master event,
// it will not for a tag push, hence the need for this handler
events.on("push", (e, p) => {
  // TODO
})

events.on("release", (e, p) => {
  // TODO
})


// Functions/Helpers

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
  var goTest = new GoJob(`${projectName}-test-integration`);
  // Enable docker so that the daemon can be used for duffle commands invoked by test-cli
  goTest.docker.enabled = true;

  // TODO: create k8s secret on infra cluster, supply appropriate name/key below
  // ALTERNATIVELY, if we save as secret in Azure Key Vault, can fetch and use.
  // This might be preferred as not tied to a particular k8s cluster, etc.
  goTest.env = {
    kubeconfig: {
      secretKeyRef: {
        name: "porter-kubeconfig",
        key: "kubeconfig"
      }
    }
  };

  goTest.tasks.push(
    "apt-get update && apt-get install -y jq",
    "mkdir -p ${HOME}/.kube",
    "echo ${kubeconfig} > ${HOME}/.kube/config",
    "export KUBECONFIG=${HOME}/.kube/config",
    "make bin/duffle-linux-amd64 test-cli"
  );

  return goTest;
}

function publish(e, p) {
  var goPublish = new GoJob(`${projectName}-publish`);

  // TODO: we could/should refactor so that this job shares a mount with the xbuild job above,
  // to remove the need of re-xbuilding before publishing

  // TODO: add 'azureStorageConnectiontring' secret to secrets section of brigade project
  goPublish.tasks.push(
    `export AZURE_STORAGE_CONNECTION_STRING=${project.secrets.azureStorageConnectiontring}`,
    "make xbuild-all publish"
  )

  return goPublish;
}

// Here we can add additional Check Runs, which will run in parallel and
// report their results independently to GitHub
function runSuite(e, p) {
  runBuild(e, p).catch(e => {console.error(e.toString())});
  runXBuild(e, p).catch(e => {console.error(e.toString())});
  runTests(e, p).catch(e => {console.error(e.toString())});
  runIntegrationTests(e, p).catch(e => {console.error(e.toString())});

  // if master, runPublish
  // TODO
}

// TODO: reduce duplication with a run<action> helper that the functions below can use

// runBuild is a Check Run that is ran as part of a Checks Suite
function runBuild(e, p) {
  // Create Notification object (which is just a Job to update GH using the Checks API)
  var note = new Notification(`build`, e, p);
  note.conclusion = "";
  note.title = "Run Go Build";
  note.summary = "Running the go build for " + e.revision.commit;
  note.text = "Ensuring the build completes successfully"

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(build(e, p), note)
}

// runBuild is a Check Run that is ran as part of a Checks Suite
function runXBuild(e, p) {
  // Create Notification object (which is just a Job to update GH using the Checks API)
  var note = new Notification(`xbuild`, e, p);
  note.conclusion = "";
  note.title = "Run Go Build - Cross-Compile";
  note.summary = "Running the go cross-compile build for " + e.revision.commit;
  note.text = "Ensuring the cross-compile build completes successfully"

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(xbuild(e, p), note)
}

// runTests is a Check Run that is ran as part of a Checks Suite
function runTests(e, p) {
  console.log("Check requested")

  // Create Notification object (which is just a Job to update GH using the Checks API)
  var note = new Notification(`tests`, e, p);
  note.conclusion = "";
  note.title = "Run Unit Tests";
  note.summary = "Running the unit tests pass for " + e.revision.commit;
  note.text = "Ensuring all unit tests pass."

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(test(e, p), note)
}

// runIntegrationTests is a Check Run that is ran as part of a Checks Suite
function runIntegrationTests(e, p) {
  console.log("Check requested")

  // Create Notification object (which is just a Job to update GH using the Checks API)
  var note = new Notification(`tests`, e, p);
  note.conclusion = "";
  note.title = "Run Integration Tests";
  note.summary = "Running the integration tests pass for " + e.revision.commit;
  note.text = "Ensuring all integration tests pass."

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(testIntegration(e, p), note)
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

// Helper to wrap a job execution between two notifications.
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

