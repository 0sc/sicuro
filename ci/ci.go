package ci

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kjk/betterguid"
)

const (
	// LogFileExt is the file extension used for log files
	LogFileExt = ".log"
)

var (
	ciDIR = filepath.Join(os.Getenv("ROOT_DIR"), "ci")
	// LogDIR is the absolute path to the CI log directory
	LogDIR = filepath.Join(ciDIR, "logs")
)

type JobDetails struct {
	LogFileName            string
	logFilePath            string
	ProjectRespositoryName string
	ProjectBranch          string
	ProjectRepositoryURL   string
	ProjectLanguage        string
	// UpdateBuildStatus func(string)
}

func init() {
	// set ci path env variable
	os.Setenv("CI_DIR", ciDIR)
}

func Run(job *JobDetails) {
	log.Println("Received job:", job)
	// job = &JobDetails{
	// 	LogFileName:            "glassbreakers/glassbreakers/master",
	// 	ProjectRepositoryURL:   "git@github.com:glassbreakers/glassbreakers",
	// 	ProjectBranch:          "master",
	// 	ProjectLanguage:        "ruby",
	// 	ProjectRespositoryName: "glassbreakers",
	// }

	job.logFilePath = filepath.Join(LogDIR, fmt.Sprintf("%s%s", job.LogFileName, LogFileExt))
	err := createDirFor(job.logFilePath)
	if err != nil {
		panic(err)
	}

	// ensure file is not being written to
	if ActiveCISession(job.logFilePath) {
		log.Println("A job is currently in progress: ", job.logFilePath)
		return
	}

	// prepare log file i.e clear file content or create new file
	if err := exec.Command("bash", "-c", "> "+job.logFilePath).Run(); err != nil {
		log.Printf("Error: %s occured while trying to clear logfile %s\n", err, job.logFilePath)
		return
	}

	job.ProjectLanguage = strings.ToLower(job.ProjectLanguage)

	log.Printf("Running job: %v\n", job)
	go runCI(job)
}

func createDirFor(fileName string) error {
	dir, file := filepath.Split(fileName)
	log.Printf("Making dir: %s for file: %s\n", dir, file)
	return os.MkdirAll(dir, 0755)
}

func runCI(job *JobDetails) {
	logFile, err := os.OpenFile(job.logFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Printf("Error %s occured while opening log file: %s\n", err, job.logFilePath)
		return
	}

	defer logFile.Close()
	// job.UpdateBuildStatus("pending")

	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s '%s' %s", filepath.Join(ciDIR, "run.sh"), prepareEnvVars(job), job.ProjectLanguage))
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err = cmd.Run()

	msg := "Test completed successfully"
	log.Println("Exit code: ", err)
	if err != nil {
		msg = fmt.Sprintf("Test failed with exit code: %s", err)
	}

	// job.UpdateBuildStatus(status)
	logFile.WriteString(fmt.Sprintf("<h4>%s</h4>", msg))
	logFile.WriteString(fmt.Sprintf("<p><a href='/run?repo=%s'>Rebuild</a><p>", job.LogFileName))
}

// ActiveCISession returns true if a ci session is active
// it returns false otherwise
func ActiveCISession(logFile string) bool {
	cmd := exec.Command("lsof", logFile)
	return cmd.Run() == nil
}

func prepareEnvVars(job *JobDetails) (vars string) {
	vars = fmt.Sprintf("%s -e %s=%s", vars, "PROJECT_BRANCH", job.ProjectBranch)
	vars = fmt.Sprintf("%s -e %s=%s", vars, "PROJECT_REPOSITORY_URL", job.ProjectRepositoryURL)
	vars = fmt.Sprintf("%s -e %s=%s", vars, "PROJECT_REPOSITORY_NAME", job.ProjectRespositoryName)
	vars = fmt.Sprintf("%s -e %s=%s", vars, "PROJECT_LANGUAGE", job.ProjectLanguage)
	vars = fmt.Sprintf("%s -e %s=%s", vars, "REDIS_URL", "redis://redis:6379")
	vars = fmt.Sprintf("%s -e %s=%s", vars, "MONGODB_URL", "mongodb://mongodb:27017")
	vars = fmt.Sprintf("%s -e %s=%s", vars, "DATABASE_URL", "postgres://postgres@postgres:5432/"+betterguid.New())

	return
}
