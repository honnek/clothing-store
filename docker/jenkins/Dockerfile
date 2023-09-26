FROM jenkins/jenkins:latest
#ENV JAVA_OPTS -Djenkins.install.runSetupWizard=false
ENV JENKINS_SLAVE_AGENT_PORT 50001

COPY --chown=jenkins:jenkins plugins.txt /usr/share/jenkins/ref/plugins.txt
RUN jenkins-plugin-cli -f /usr/share/jenkins/ref/plugins.txt