package commands

var (
	ignoreScripts bool
	// Ignore dependencies in the optionalDependencies field
	ignoreOptional bool
	// Only add, change or remove dependencies in the devDependencies field
	saveDev bool
	// Only add, change or remove dependencies in the dependencies field
	saveProd bool
)
