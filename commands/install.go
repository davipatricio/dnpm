package commands

import (
	"dnpm/messages"
	"dnpm/rest"
	"dnpm/structs"
	"dnpm/utils"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gookit/color"
)

func RunInstallCmd() bool {
	// Argument parsing

	// os.Args[1] will always be "add", "install" or "i" (see dnpm.go)
	installCmd := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	showEmojis := installCmd.Bool("emoji", false, "Whether to show emojis on the output.")
	showDebug := installCmd.Bool("debug", false, "Whether to show additional information on the output.")
	downloadDev := installCmd.Bool("download-dev", false, "Whether to download dev depedencies.")
	downloadOptionalDep := installCmd.Bool("download-opt", false, "Whether to download optional depedencies.")

	// Command code
	path, found, _ := utils.GetNearestPackageJSON()

	// Check if the only argument/subcommand is "install"
	if len(os.Args) == 2 {
		installCmd.Parse(os.Args[2:])
		// If we found a package.json, we should run installPackagesPresentOnPackageJSON()
		// and install all dependencies and devDependecies.
		if found {
			messages.FoundPkgInstallCmd(*showEmojis)
			os.Mkdir("node_modules", 0755)
			installPackagesPresentOnPackageJSON(path, *showEmojis, *showDebug, *downloadDev, *downloadOptionalDep)
			return false
		}
		// Tell the user that we couldn't find a package.json and recommend the use of "dnpm init"
		messages.NoPkgJSONFoundInstallCmd(*showEmojis)
		return false
	}

	// Check if we found a package.json and there are packages/arguments
	if found {
		installCmd.Parse(os.Args[2:])
		packagesArgs := installCmd.Args()
		if len(packagesArgs) < 1 {
			// If the user provide only an argument (e.g "dnpm install --emoji")
			// Tell the user no packages were provided to install
			messages.NoPkgProvidedInstallCmd(*showEmojis)
			return false
		}

		os.Mkdir("node_modules", 0755)
		// Notify the user that we are installing the requested packages
		messages.InstallingPkgsInstallCmd(*showEmojis, packagesArgs)
		installSpecificPackages(packagesArgs, false, true, *showEmojis, *showDebug, *downloadDev, *downloadOptionalDep)
		return false
	}

	// Tell the user that we couldn't find a package.json and recommend the use of "dnpm init"
	messages.NoPkgJSONFoundInstallCmd(*showEmojis)

	return false
}

// If this function is called, means that we found a package.json and no packages/arguments were provided
func installPackagesPresentOnPackageJSON(path string, showEmojis, showDebug, downloadDev, downloadOptionalDep bool) {
	// Read the package.json file
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// Parse the JSON
	var pkgJSON structs.PackageJSONFormat
	json.Unmarshal(jsonFile, &pkgJSON)

	fmt.Println(pkgJSON)

	createEmptyStoreFolder()
	createEmptyTempFolder()

	// Get the folder we store cached packages
	for pkgName, pkgVersion := range pkgJSON.Dependencies {
		pkgVersion = utils.RemovePkgVersionRange(pkgVersion)

		setAlreadyInstalling(pkgName, pkgVersion)
		// If the folder doesn't exist, we should create it
		createEmptyFolderForPkg(pkgName, pkgVersion)
		createTempFolderForPkg(pkgName)

		// Download the package
		installDebug("Downloading package "+pkgName+" ("+pkgVersion+")", false)
		installSpecificPackages([]string{pkgName + "@" + pkgVersion}, false, true, showEmojis, showDebug, downloadDev, downloadOptionalDep)
	}
}

func installSpecificPackages(packages []string, isDep, manual, showEmojis, showDebug, downloadDev, downloadOptionalDep bool) {
	startTime := int64(0)

	// avoid allocate memory
	if manual {
		startTime = time.Now().UnixMilli()
	}

	for _, rawPkgString := range packages {
		// Get the package name from the provided string e.g. "typescript@nightly" -> "typescript"
		pkgName := utils.GetPkgName(rawPkgString)
		// Get the package name from the provided string e.g. "@myorg/mypkg@nightly" -> "mypkg"
		pkgWithoutOrgName := utils.RemoveOrgName(pkgName)
		// Empty string if the package doesn't have an org
		pkgOrgName := utils.GetOrgName(pkgName)
		// Get the version from the provided string e.g. "typescript@nightly" -> "nightly"
		// e.g. "typescript" -> ""
		pkgVersion := utils.GetPkgVersionOrTag(rawPkgString)
		// Make a request to the Yarn registry requesting the package info
		d, err := rest.GetPkg(pkgName)

		if d["error"] != nil || err != nil {
			messages.PkgNotFoundInstallCmd(showEmojis, pkgName)
			continue
		}

		latestVersion := ""

		// If the package has a tag that is the same as the provided version,
		// we should install that tag instead of the version.
		if d["dist-tags"] != nil {
			distTags := d["dist-tags"].(map[string]interface{})
			latestVersion = distTags["latest"].(string)
			// If no version was provided, use the latest version
			if pkgVersion == "" {
				// Get the property latest from d.dist-tags
				pkgVersion = latestVersion
			} else if distTags[pkgVersion] != nil {
				// Get the version of the tag
				pkgVersion = distTags[pkgVersion].(string)
			}

			// If the requested version of a dependency is not available
			// we should try to install the latest version
			if d["versions"].(map[string]interface{})[pkgVersion] == nil && isDep {
				pkgVersion = latestVersion
				if !isAlreadyInstalling(pkgName, pkgVersion) {
					installDebug(" !!! Depedency "+pkgName+" does not has version "+pkgVersion+". Using the latest version!", showDebug)
				}
			}
		}

		if isAlreadyInstalling(pkgName, pkgVersion) || d["versions"].(map[string]interface{})[pkgVersion] == nil {
			continue
		}

		// Verify if the package is already cached
		depCached := isPkgAlreadyCached(pkgName, pkgVersion)
		if !isDep {
			addDepsToPackageJSON(pkgName, pkgVersion, "dependencies")
		} else {
			addDepsToPackageJSON(pkgName, pkgVersion, "devDependencies")
		}

		createEmptyStoreFolder()
		createEmptyTempFolder()

		setAlreadyInstalling(pkgName, pkgVersion)

		// If the folder doesn't exist, we should create it
		createEmptyFolderForPkg(pkgName, pkgVersion)
		createTempFolderForPkg(pkgName)

		versionData := d["versions"].(map[string]interface{})[pkgVersion].(map[string]interface{})
		downloadUrl := versionData["dist"].(map[string]interface{})["tarball"].(string)
		// "@types\node\18.0.0\node" -> "@types\node\18.0.0"
		pathWithoutDuplicatedName := utils.RemoveLastSubstring(utils.GetStoreDir()+"/"+pkgName+"/"+pkgVersion, "/"+pkgWithoutOrgName)
		pathWithoutDuplicatedName = utils.RemoveLastSubstring(pathWithoutDuplicatedName, "\\"+pkgWithoutOrgName)

		ch := make(chan bool)
		go func() {
			if !depCached {
				installDebug("Downloading package "+pkgName+" ("+pkgVersion+")", showDebug)
				// Download the tgz to the temp folder
				rest.DownloadPkgTgz(downloadUrl, utils.GetTempDir()+"/"+pkgName+"/"+pkgVersion+".tgz")
				installDebug("Extracting package "+pkgName+" ("+pkgVersion+")", showDebug)
				// Extract the tgz to the store folder
				utils.DecompressTgz(utils.GetTempDir()+"/"+pkgName+"/"+pkgVersion+".tgz", pathWithoutDuplicatedName)
				go setAlreadyCached(pkgName, pkgVersion)
				go removeAlreadyInstalling(pkgName, pkgVersion)

				// Remove the temp tgz
				go os.Remove(utils.GetTempDir() + "/" + pkgName + "/" + pkgVersion + ".tgz")
				// Install the package to the node_modules folder
				installDebug("Installing package "+pkgName+" ("+pkgVersion+")", showDebug)
			} else {
				installDebug("Package "+pkgName+" ("+pkgVersion+") is already cached. Skipping download.", showDebug)
			}
			ch <- true
			installToNodeModules(pkgOrgName, pkgName, utils.GetStoreDir()+"/"+pkgName+"/"+pkgVersion+"/package")
		}()
		<-ch

		done := make(chan bool, 3)

		// Check if there are dependencies
		deps, ok := versionData["dependencies"].(map[string]interface{})
		if ok {
			go loopAndDownloadDeps(deps, showEmojis, showDebug, done)
		} else {
			done <- true
		}

		devDeps, ok := versionData["devDependencies"].(map[string]interface{})
		if downloadDev && ok && !isDep {
			go loopAndDownloadDeps(devDeps, showEmojis, showDebug, done)
		} else {
			done <- true
		}

		optDeps, ok := versionData["optionalDependencies"].(map[string]interface{})
		if downloadOptionalDep && ok && !isDep {
			go loopAndDownloadDeps(optDeps, showEmojis, showDebug, done)
		} else {
			done <- true
		}

		<-done
		<-done
		<-done
	}

	if manual {
		endTime := time.Now().UnixMilli()
		// We should check this so we don't spam the output
		// Saying which packages were downloaded
		messages.DoneInstallCmd(showEmojis, endTime-startTime)
	}
}

func addDepsToPackageJSON(pkgName, pkgVersion, depType string) {
	dir, _, err := utils.GetNearestPackageJSON()
	if err != nil {
		return
	}

	// Read the package.json file
	packageJSON, err := ioutil.ReadFile(dir)
	if err != nil {
		return
	}

	// Parse the package.json file
	var pkgJSON structs.PackageJSONFormat
	err = json.Unmarshal(packageJSON, &pkgJSON)
	if err != nil {
		return
	}

	if pkgJSON.Dependencies == nil {
		pkgJSON.Dependencies = make(map[string]string)
	}

	if pkgJSON.DevDependencies == nil {
		pkgJSON.DevDependencies = make(map[string]string)
	}

	// Add the dependency to the package.json file
	if depType == "dependencies" {
		pkgJSON.Dependencies[pkgName] = pkgVersion
	}
	if depType == "devDependencies" {
		pkgJSON.DevDependencies[pkgName] = pkgVersion
	}

	// Convert the package.json file to a string
	packageJSON, err = json.Marshal(pkgJSON)
	if err != nil {
		return
	}

	// Write the package.json file
	err = ioutil.WriteFile(dir, packageJSON, 0644)
	if err != nil {
		return
	}
}

func loopAndDownloadDeps(deps map[string]interface{}, showEmojis, showDebug bool, ch chan bool) {
	// Loop through each dependencies
	for depName, depVer := range deps {
		ch2 := make(chan bool, 1)
		cleanDepVer := utils.RemovePkgVersionRange(depVer.(string))
		go downloadDeps(depName, cleanDepVer, showEmojis, showDebug, ch2)
		<-ch2
	}
	ch <- true
}

func downloadDeps(depName, depVer string, showEmojis, showDebug bool, ch chan bool) {
	installDebug("Verifying if dependency "+depName+" ("+depVer+") is cached", showDebug)
	// If the dependency is not cached, we should download it
	if !isAlreadyInstalling(depName, depVer) && !isPkgAlreadyCached(depName, depVer) {
		installDebug("Dependency "+depName+" is not cached.\nDownloading dependency "+depName+" ("+depVer+")", showDebug)
		// Call this function again to download the dependency
		// So we don't have duplicated code
		installSpecificPackages([]string{depName + "@" + depVer}, true, false, showEmojis, showDebug, false, false)
	}
	ch <- true
}

// Check if a package is installing
func isAlreadyInstalling(pkg, version string) bool {
	// Get the folder we store cached packages
	dir := utils.GetStoreDir()
	_, err := os.Stat(dir + "/" + pkg + "/" + version + "/package/.dnpm-installing")
	return err == nil
}

func setAlreadyInstalling(pkg, version string) {
	// Get the folder we store cached packages
	dir := utils.GetStoreDir()
	// Create the file
	ioutil.WriteFile(dir+"/"+pkg+"/"+version+"/package/.dnpm-installing", []byte(""), 0644)
}

func removeAlreadyInstalling(pkg, version string) {
	// Get the folder we store cached packages
	dir := utils.GetStoreDir()
	// Remove the file
	os.Remove(dir + "/" + pkg + "/" + version + "/package/.dnpm-installing")
}

// Check if a package already was cached
func isPkgAlreadyCached(pkg, version string) bool {
	storeDir := utils.GetStoreDir()
	_, err := os.Stat(storeDir + "/" + pkg + "/" + version + "/package")
	_, err2 := os.Stat(storeDir + "/" + pkg + "/" + version + "/package/.dnpm-download-complete")
	return err == nil && err2 == nil
}

func setAlreadyCached(pkg, version string) {
	// Get the folder we store cached packages
	dir := utils.GetStoreDir()
	// Create the file
	ioutil.WriteFile(dir+"/"+pkg+"/"+version+"/package/.dnpm-download-complete", []byte(""), 0644)
}

// Create a cached folder for a package
func createEmptyFolderForPkg(pkg, version string) {
	// Get the folder we store cached packages
	dir := utils.GetStoreDir()
	// Verify if the package is already cached
	depCached := isPkgAlreadyCached(pkg, version)
	if !depCached {
		// If the package is not cached, we should create the folder
		err := os.MkdirAll(dir+"/"+pkg+"/"+version+"/package", 0755)
		if err != nil {
			panic(err)
		}
	}
}

func createEmptyStoreFolder() {
	storeDir := utils.GetStoreDir()
	// Verify if the store folder exists
	_, err := os.Stat(storeDir)
	if err != nil {
		err = os.MkdirAll(storeDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

// Create a temp folder for a package that is being downloaded
func createEmptyTempFolder() {
	tempDir := utils.GetTempDir()
	// Verify if the temp folder exists
	_, err := os.Stat(tempDir)
	if err != nil {
		err = os.MkdirAll(tempDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func createTempFolderForPkg(pkg string) {
	// Get the folder we store cached packages
	dir := utils.GetTempDir()
	// If the package is not cached, we should create the folder
	os.MkdirAll(dir+"/"+pkg, 0755)
}

func installToNodeModules(org, pkg, dir string) {
	if org == "" {
		utils.CreateSymlink(dir, "node_modules/"+pkg)
	} else {
		_, err := os.Stat("node_modules/" + org)
		if err != nil {
			os.MkdirAll("node_modules/"+org, os.ModePerm)
		}
		utils.CreateSymlink(dir, "node_modules/"+pkg)
	}
}

func installDebug(info string, showDebug bool) {
	if showDebug {
		color.Gray.Println(info)
	}
}
