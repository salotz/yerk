package models

// Tagged union of yerk wire resource kinds. Discriminator is kind.
#Resource:
	#Catalog |
	#Project |
	#ProjectStatus |
	#ReplicaStatus |
	#ProjectInfo |
	#ReplicaInfo |
	#ConfigResolve |
	#ToolConfig |
	#ToolContext |
	#DirContext |
	#EnvVars
