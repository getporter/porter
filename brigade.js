const { events, Job, Group } = require("brigadier");

// **********************************************
// Globals
// **********************************************

const projectName = "porter";
const projectOrg = "deislabs";

// **********************************************
// Event Handlers
// **********************************************

events.on("check_suite:requested", runSuite)
events.on("check_suite:rerequested", runSuite)
events.on("check_run:rerequested", runSuite)

events.on("exec", (e, p) => {
  Group.runAll([
    build(e, p),
    xbuild(e, p),
    test(e, p),
    testIntegration(e, p)
  ]);
})

// Although a GH App will trigger 'check_suite:requested' on a push to master event,
// it will not for a tag push, hence the need for this handler
events.on("push", (e, p) => {
  if (e.revision.ref.includes("refs/heads/master") || e.revision.ref.startsWith("refs/tags/")) {
    publish(e, p).run();
  }
})

events.on("publish", (e, p) => {
  publish(e, p).run();
})

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
  var goTest = new GoJob(`${projectName}-test-integration`);
  // Enable docker so that the daemon can be used for duffle commands invoked by test-cli
  goTest.docker.enabled = true;

  goTest.env.kubeconfig = {
    secretKeyRef: {
      name: "porter-kubeconfig",
      key: "kubeconfig"
    }
  };

  // Install docker cli
  goTest.tasks.push(
    "apt-get update && apt-get install -y jq apt-transport-https ca-certificates curl gnupg2 software-properties-common",
    "curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -",
    `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"`,
    "apt-get update && apt-get install -y docker-ce"
  )

  goTest.tasks.push(
    "mkdir -p ${HOME}/.kube",
    'echo "${kubeconfig}" > ${HOME}/.kube/config',
    `docker login ${p.secrets.dockerhubRegistry} \
      -u ${p.secrets.dockerhubUsername} \
      -p ${p.secrets.dockerhubPassword}`,
    `REGISTRY=${p.secrets.dockerhubOrg} make bin/duffle-linux-amd64 test-cli`
  );

  return goTest;
}

function publish(e, p) {
  var goPublish = new GoJob(`${projectName}-publish`);

  // Install az cli
  goPublish.tasks.push(
    `apt-get update && apt-get install apt-transport-https lsb-release software-properties-common dirmngr -y`,
    `AZ_REPO=$(lsb_release -cs) && \
      echo "AZ_REPO = $AZ_REPO" && \
      echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $AZ_REPO main" | \
        tee /etc/apt/sources.list.d/azure-cli.list`,
    `apt-key --keyring /etc/apt/trusted.gpg.d/Microsoft.gpg adv \
      --keyserver packages.microsoft.com \
      --recv-keys BC528686B50D79E339D3721CEB3E94ADBE1229CF`,
    `apt-get update && apt-get install azure-cli -y`
  )

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
  Group.runAll([
    new CheckRun("build", build(e, p), e, p),
    new CheckRun("xbuild", xbuild(e, p), e, p),
    new CheckRun("tests", test(e, p), e, p),
    new CheckRun("integrationTests", testIntegration(e, p), e, p)
  ]).catch(e => {
    console.error(e.toString());
  });
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

// CheckRun returns a GitHub Check Run that can be run as part of a GitHub Checks Suite
class CheckRun {
  constructor(action, actionFunc, e, p) {
    this.notification = new Notification(action, e, p);
    this.notification.conclusion = "";
    this.notification.title = `Run ${action}`;
    this.notification.summary = `Running ${action} for ${e.revision.commit}`;
    this.notification.text = `Ensuring ${action} completes successfully`

    this.actionFunc = actionFunc;
  }

  run() {
    return notificationWrap(this.actionFunc, this.notification);
  }
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