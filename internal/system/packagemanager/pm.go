package packagemanager

type Package struct {
	Name              string
	Version           string
	NativePackageName map[string]string
	SystemPackage     bool
	Library           bool
	Optional          bool
}

type packagemap = map[string][]*Package

type PackageManager interface {
	Name() string
	Packages() packagemap
	PackageInstalled(pkg *Package) (bool, error)
	PackageAvailable(pkg *Package) (bool, error)
	InstallCommand(pkg *Package) string
}

func GenerateUniversalPackages() packagemap {
	return packagemap{
		// Entwicklungstools
		"git": {
			{
				Name:          "git",
				SystemPackage: true,
				NativePackageName: map[string]string{
					"apt":      "git",
					"dnf":      "git",
					"pacman":   "git",
					"zypper":   "git",
					"homebrew": "git",
					"choco":    "git",
				},
			},
		},
		// Java-Entwicklung
		"openjdk": {
			{
				Name:          "openjdk-17",
				SystemPackage: true,
				NativePackageName: map[string]string{
					"apt":      "openjdk-17-jdk",
					"dnf":      "java-17-openjdk-devel",
					"pacman":   "jdk17-openjdk",
					"zypper":   "java-17-openjdk-devel",
					"homebrew": "openjdk@17",
					"choco":    "openjdk-17",
				},
			},
		},
		// Container-Technologien
		"podman": {
			{
				Name:          "podman",
				SystemPackage: true,
				Optional:      true,
				NativePackageName: map[string]string{
					"apt":      "podman",
					"dnf":      "podman",
					"pacman":   "podman",
					"zypper":   "podman",
					"homebrew": "podman",
					"choco":    "podman",
				},
			},
		},
		// Entwicklungsumgebungen
		"vscode": {
			{
				Name:          "vscode",
				SystemPackage: false,
				Optional:      true,
				NativePackageName: map[string]string{
					"apt":      "code",
					"dnf":      "code",
					"pacman":   "code",
					"zypper":   "code",
					"homebrew": "visual-studio-code",
					"choco":    "vscode",
				},
			},
		},
		// Build-Tools
		"maven": {
			{
				Name:          "maven",
				SystemPackage: true,
				NativePackageName: map[string]string{
					"apt":      "maven",
					"dnf":      "maven",
					"pacman":   "maven",
					"zypper":   "maven",
					"homebrew": "maven",
					"choco":    "maven",
				},
			},
		},
		"gradle": {
			{
				Name:          "gradle",
				SystemPackage: true,
				Optional:      true,
				NativePackageName: map[string]string{
					"apt":      "gradle",
					"dnf":      "gradle",
					"pacman":   "gradle",
					"zypper":   "gradle",
					"homebrew": "gradle",
					"choco":    "gradle",
				},
			},
		},
		// Compiler und Entwicklungstools
		"gcc": {
			{
				Name:          "gcc",
				SystemPackage: true,
				NativePackageName: map[string]string{
					"apt":      "build-essential",
					"dnf":      "gcc",
					"pacman":   "gcc",
					"zypper":   "gcc",
					"homebrew": "gcc",
					"choco":    "mingw-gcc",
				},
			},
		},
	}
}

type Dependency struct {
	Name           string
	PackageName    string
	Installed      bool
	InstallCommand string
	Version        string
	Optional       bool
	External       bool
}

type DependencyList []*Dependency

func (d DependencyList) InstallAllRequiredCommand() string {
	result := ""
	for _, dependency := range d {
		if !dependency.Installed && !dependency.Optional {
			result += "  - " + dependency.Name + ": " + dependency.InstallCommand + "\n"
		}
	}

	return result
}

func (d DependencyList) InstallAllOptionalCommand() string {
	result := ""
	for _, dependency := range d {
		if !dependency.Installed && dependency.Optional {
			result += "  - " + dependency.Name + ": " + dependency.InstallCommand + "\n"
		}
	}

	return result
}
