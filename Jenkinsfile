#!/usr/bin/env groovy

properties([
	buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', numToKeepStr: '5')),
	[$class: 'ScannerJobProperty', doNotScan: true], 
	[$class: 'GithubProjectProperty', projectUrlStr: 'http://github.so.kadaster.nl/PDOK/capabilities-proxy/']
])

final String appName = 'capabilities-proxy'
final String applicationContainerName = appName.toUpperCase() + '-DOCKER'
final String appVersion = appName.toUpperCase() + "-${env.BUILD_NUMBER}".toString()
final String job = "${env.JOB_NAME} #${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"

final werkomgevingDockerHostMap = [:]
werkomgevingDockerHostMap['prod'] = ['csu420.cs.kadaster.nl', 'csu395.cs.kadaster.nl']

try {	
	timeout(60) {
		stage(Globals.STAGE_BUILD) {
			node {
				checkout scm
				withEnv([
						"DOCKER_HOST=tcp://inu342.in.kadaster.nl:2376",
						"DOCKER_CERT_PATH=${env.JENKINS_HOME}/.docker/inu342",
						"DOCKER_TLS_VERIFY=1"
				]) {
					docker.withRegistry("http://${env.DOCKER_PROD_URL}") {
						docker.build('pdok/capabilities-proxy').push(appVersion)
                        docker.build('pdok/capabilities-proxy').push("latest")
					}
				}
			}
		}
	}
	
	
	new Dockerapp().deployToDocker(werkomgevingDockerHostMap, null, appName, applicationContainerName, appVersion, "capabilities-proxy", "9001", true)
	
} catch (org.jenkinsci.plugins.workflow.steps.FlowInterruptedException exception) {
    errorMessage = "${exception}"
    echo "Caught: ${exception}"

    if (exception.result == Result.ABORTED) {
        currentBuild.result = "Aborted"
    } else {
        currentBuild.result = "Failure"
        new Slack().sendMessage {
            channel = '#builds'
            color = 'danger'
            message = "Failure for ${job} :angry: \n`${errorMessage}`"
        }
    }
} catch (exception) {
    currentBuild.result = "Failure"
    errorMessage = "${exception}"
    echo "Caught: ${errorMessage}"

    new Slack().sendMessage {
        channel = '#builds'
        color = 'danger'
        message = "Failure for ${job} :angry: \n`${errorMessage}`"
    }
} finally {
    timeout(60) {
        node() {
            echo "Cleanup workspace"
            deleteDir()
        }
    }
}