const { events, Job, Group } = require("brigadier");
const { TestingJobGenerator, ImageJobGenerator, NotificationJobGenerator, BuildStatus, Debugger } = require("./spotahome");
    
const buildStatusSuccess = "success";
    
const eventTypePush = "push";
const eventTypeExec = "exec";
const eventTypePullRequest = "pull_request";
const eventTypeAfter = "after";
const eventTypeError = "error";


/**
 * fullBuild is the most complete kind of build on the pipeline, 
 * it runs the unit tests, build the release image and pushes to
 * a registry.
 */
function fullBuild(e, project) {

    let ijg = new ImageJobGenerator(e, project);
    env = {
        IMAGE_VERSION: e.revision.commit,
        PUSH_IMAGE: "true",
        TARGET: "mandrill-prometheus-exporter"
    }
    let buildJob = ijg.dockerImageEnvironment("./hack/build-image.sh", env);

    // Run the tests and then build release.
    Group.runEach([buildJob])
}

/**
 * debugFullBuild is like fullbuild but prints the received event information.
 * useful while developing. WARNING: project is not printed because it has secrets
 * and could be leaked by accident.
 */
function debugFullBuild(e, project) {
    let d = new Debugger(e, project);
    d.debugEvent(e);
    fullBuild(e, project);
}

     
/**
 * buildStatusToGithub will put the correct build status on github.
 */
function buildStatusToGithub(e, project) {
    // Only set state of build when is a push or PR.
    if ([eventTypePush, eventTypePullRequest].includes(e.cause.event.type)) {
        // Set correct status
        let state = BuildStatus.Failure;
        if (e.cause.trigger == buildStatusSuccess) {
            state = BuildStatus.Success;
        }
        // Set the status on github.
        let njg = new NotificationJobGenerator(e, project);
        njg.githubStatus(state).run();
        // Set if the commit es deployable.
        njg.githubSetToiletDeployable(state).run();
    } else {
        console.log(`Build finished with ${e.cause.trigger} state`);
    }
}


events.on(eventTypePush, fullBuild);
events.on(eventTypeExec, debugFullBuild);

// Final events after build (failure or success).
events.on(eventTypeAfter, buildStatusToGithub);
events.on(eventTypeError, buildStatusToGithub);
