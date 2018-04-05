package jenkins

import (
	"github.com/jenkins-x/jx/pkg/gits"
)

func CreateFolderXml(folderUrl string, name string) string {
	return `<?xml version='1.0' encoding='UTF-8'?>
<com.cloudbees.hudson.plugins.folder.Folder plugin="cloudbees-folder@6.2.1">
  <actions>
    <io.jenkins.blueocean.service.embedded.BlueOceanUrlAction plugin="blueocean-rest-impl@1.3.3">
      <blueOceanUrlObject class="io.jenkins.blueocean.service.embedded.BlueOceanUrlObjectImpl">
        <mappedUrl>blue/organizations/jenkins</mappedUrl>
        <modelObject class="com.cloudbees.hudson.plugins.folder.Folder" reference="../../../.."/>
      </blueOceanUrlObject>
    </io.jenkins.blueocean.service.embedded.BlueOceanUrlAction>
  </actions>
  <description></description>
  <properties>
    <org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig plugin="pipeline-model-definition@1.2.4">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons@1.9"/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig>
  </properties>
  <folderViews class="com.cloudbees.hudson.plugins.folder.views.DefaultFolderViewHolder">
    <views>
      <hudson.model.AllView>
        <owner class="com.cloudbees.hudson.plugins.folder.Folder" reference="../../../.."/>
        <name>All</name>
        <filterExecutors>false</filterExecutors>
        <filterQueue>false</filterQueue>
        <properties class="hudson.model.View$PropertyList"/>
      </hudson.model.AllView>
    </views>
    <tabBar class="hudson.views.DefaultViewsTabBar"/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="com.cloudbees.hudson.plugins.folder.icons.StockFolderIcon"/>
</com.cloudbees.hudson.plugins.folder.Folder>
`
}

func createBranchSource(info *gits.GitRepositoryInfo, gitProvider gits.GitProvider, credentials string, branches string) string {
	credXml := ""
	if credentials != "" {
		credXml = `		  <credentialsId>` + credentials + `</credentialsId>
`
	}
	if gitProvider.IsGitHub() {
		serverXml := ""
		ghp, ok := gitProvider.(*gits.GitHubProvider)
		if ok {
			u := ghp.GetEnterpriseApiURL()
			if u != "" {
				serverXml = `		  <apiUri>` + u + `</apiUri>
`
			}
		}
		return `
	    <source class="org.jenkinsci.plugins.github_branch_source.GitHubSCMSource" plugin="github-branch-source@2.3.1">
		  <id>b50ee5d4-cb45-42de-9140-d79330bab9ac</id>` + credXml + serverXml + `
		  <repoOwner>` + info.Organisation + `</repoOwner>
		  <repository>` + info.Name + `</repository>
		  <traits>
			<org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait>
			  <strategyId>1</strategyId>
			</org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait>
			<org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait>
			  <strategyId>1</strategyId>
			</org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait>
			<org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait>
			  <strategyId>1</strategyId>
			  <trust class="org.jenkinsci.plugins.github_branch_source.ForkPullRequestDiscoveryTrait$TrustContributors"/>
			</org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait>
			<jenkins.scm.impl.trait.RegexSCMHeadFilterTrait plugin="scm-api@2.2.6">
			  <regex>` + branches + `</regex>
			</jenkins.scm.impl.trait.RegexSCMHeadFilterTrait>
		  </traits>
		</source>
`
	}
	if gitProvider.IsGitea() {
		return `
	    <source class="org.jenkinsci.plugin.gitea.GiteaSCMSource" plugin="gitea@1.0.5">
          <id>db44ccb9-31c0-4b78-8989-614af3a87b9f</id>` + credXml + `
          <serverUrl>` + info.HostURLWithoutUser() + `</serverUrl>
          <repoOwner>` + info.Organisation + `</repoOwner>
		  <repository>` + info.Name + `</repository>
          <traits>
            <org.jenkinsci.plugin.gitea.BranchDiscoveryTrait>
              <strategyId>1</strategyId>
            </org.jenkinsci.plugin.gitea.BranchDiscoveryTrait>
            <org.jenkinsci.plugin.gitea.OriginPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
            </org.jenkinsci.plugin.gitea.OriginPullRequestDiscoveryTrait>
            <org.jenkinsci.plugin.gitea.ForkPullRequestDiscoveryTrait>
              <strategyId>1</strategyId>
              <trust class="org.jenkinsci.plugin.gitea.ForkPullRequestDiscoveryTrait$TrustContributors"/>
            </org.jenkinsci.plugin.gitea.ForkPullRequestDiscoveryTrait>
			<jenkins.scm.impl.trait.RegexSCMHeadFilterTrait plugin="scm-api@2.2.6">
			  <regex>` + branches + `</regex>
			</jenkins.scm.impl.trait.RegexSCMHeadFilterTrait>
		  </traits>
		</source>
`
	}
	return `
<source class="jenkins.plugins.git.GitSCMSource" plugin="git@3.7.0">
  <id>3ee777bd-6590-4b97-ac65-1ab01e7062ad</id>
  <remote>` + info.URL + `</remote>
` + credXml + `
<traits>
	<jenkins.plugins.git.traits.BranchDiscoveryTrait/>
  </traits>
</source>
<strategy class="jenkins.branch.DefaultBranchPropertyStrategy">
  <properties class="empty-list"/>
</strategy>
`
}

func CreateMultiBranchProjectXml(info *gits.GitRepositoryInfo, gitProvider gits.GitProvider, credentials string, branches string, jenkinsfile string) string {
	return `<?xml version='1.0' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch@2.16">
  <actions/>
  <description></description>
  <properties>
	<org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig plugin="pipeline-model-definition@1.2.4">
	  <dockerLabel></dockerLabel>
	  <registry plugin="docker-commons@1.9"/>
	</org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api@2.0.15">
	<owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
	<com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder@6.2.1">
	  <nonRecursive>false</nonRecursive>
	</com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api@2.0.15">
	<owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
  <orphanedItemStrategy class="com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy" plugin="cloudbees-folder@6.2.1">
	<pruneDeadBranches>true</pruneDeadBranches>
	<daysToKeep>-1</daysToKeep>
	<numToKeep>-1</numToKeep>
  </orphanedItemStrategy>
  <triggers/>
  <disabled>false</disabled>
  <sources class="jenkins.branch.MultiBranchProject$BranchSourceList" plugin="branch-api@2.0.15">
	<data>
	  <jenkins.branch.BranchSource>
` + createBranchSource(info, gitProvider, credentials, branches) + `
		<strategy class="jenkins.branch.DefaultBranchPropertyStrategy">
		  <properties class="empty-list"/>
		</strategy>
	  </jenkins.branch.BranchSource>
	</data>
	<owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </sources>
  <factory class="org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory">
	<owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
	<scriptPath>` + jenkinsfile + `</scriptPath>
  </factory>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>
`
}
