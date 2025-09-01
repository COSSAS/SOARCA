package jupyter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	sshcapability "soarca/pkg/core/capability/ssh"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"strings"

	"github.com/google/uuid"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
)

type JupyterCapability struct {
	cached_notebooks map[int64]string
}

var component = reflect.TypeOf(JupyterCapability{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Trace, "", logger.Json)
}

func New() JupyterCapability {
	return JupyterCapability{
		cached_notebooks: make(map[int64]string),
	}
}

func newRandomFilename() string {
	return "notebook--" + uuid.NewString() + ".ipynb"
}

func newSSHClient(context capability.Context) (*ssh.Session, *ssh.Client, error) {
	err := sshcapability.CheckSshAuthenticationInfo(context.Authentication)
	if err != nil {
		log.Error(err)
		return nil, nil, err
	}
	config, err := sshcapability.GetConfig(context.Authentication)
	if err != nil {
		return nil, nil, err
	}
	return sshcapability.GetSession(config, context.Target)
}

func (capability *JupyterCapability) ExecuteB64(encoded string, metadata execution.Metadata, context capability.Context) (cacao.Variables, error) {
	// decoded, err := base64.StdEncoding.DecodeString(encoded)
	// if err != nil {
	// 	log.Error("Could not decode b64 notebook")
	// 	return cacao.NewVariables(), err
	// }
	return ExecuteContainer(encoded, metadata, context)
	// return ExecuteRemote(decoded, metadata, context)
}

// 	if err := os.Mkdir("lab", 0666); err != nil && !os.IsExist(err) {
// 		return "", err
// 	}
// 	filename := path.Join("lab", "notebook-"+uuid.NewString()+".ipynb")
// 	err := os.WriteFile(filename, bytes, 0666)
// 	return filename, err
// }

func ExecuteContainer(encoded string, metadata execution.Metadata, ctxt capability.Context) (cacao.Variables, error) {
	// notebook_filename, err := writeTempFile(decoded)
	// if err != nil {
	// 	return cacao.NewVariables(), err
	// }
	ctx := context.Background()
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace("jupy"))
	if err != nil {
		return cacao.NewVariables(), err
	}
	existing_containers, err := client.Containers(ctx)
	if err != nil {
		return cacao.NewVariables(), err
	}
	var container containerd.Container
	found := false
	for _, cont := range existing_containers {
		if cont.ID() == "ju" {
			container = cont
			found = true
			break
		}
	}
	if !found {
		image_reader, err := os.Open("./jupy.tar.gz")
		if err != nil {
			return cacao.NewVariables(), err
		}
		images, err := client.Import(ctx, image_reader)
		if err != nil {
			return cacao.NewVariables(), err
		}
		if len(images) != 1 {
			return cacao.NewVariables(), errors.New("did not load an image")
		}

		container_image := containerd.NewImage(client, images[0])
		// img_spec, err := oci.GenerateSpec(ctx, client, oci.WithImageConfig(container_image))
		// if err != nil {
		// 	return cacao.NewVariables(), err
		// }

		container, err = client.NewContainer(ctx, "ju", containerd.WithImage(container_image), containerd.WithNewSnapshot("ju-snapshot", container_image))
		if err != nil {
			return cacao.NewVariables(), err
		}
	}
	// cwd, err := os.Getwd()
	// if err != nil {
	// 	return cacao.NewVariables(), errors.Join(errors.New("cwd:"), err)
	// }
	// rootfs_dir := path.Join(cwd, "lab_fs")
	// rootfs_contents, err := os.ReadDir(rootfs_dir)
	// if err != nil {
	// 	os.Mkdir(rootfs_dir, 0755)
	// } else if len(rootfs_contents) != 0 {
	// 	return cacao.NewVariables(), errors.New("Container Root fs is not empty")
	// }
	// nb_mount := mount.Mount{Source: rootfs_dir, Type: "bind", Options: []string{"rbind", "ro"}}
	// log.Info("Mounting ", nb_mount.Source, " to ", nb_mount.Target)
	// if err := nb_mount.Mount(path.Join(cwd, "lab")); err != nil {
	// 	return cacao.NewVariables(), errors.Join(errors.New("mount:"), err)
	// }
	task, err := container.NewTask(ctx, cio.NewCreator()) //, containerd.WithRootFS([]mount.Mount{nb_mount}))
	if err != nil {
		return cacao.NewVariables(), errors.Join(errors.New("task creation:"), err)
	}
	full_cmd := "jupyter nbconvert --to notebook --execute --stdin --stdout 2>/dev/null"

	proc := specs.Process{Args: strings.Split(full_cmd, " ")}
	stdin := strings.NewReader(encoded)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	task_proc, err := task.Exec(ctx, "", &proc, cio.NewCreator(cio.WithStreams(stdin, stdout, stderr)))
	if err != nil {
		log.Error("stderr output: ", stderr.String())
		return cacao.NewVariables(), errors.Join(errors.New("task exe:"), err)
	}
	status_chan, err := task_proc.Wait(ctx)
	if err != nil {
		return cacao.NewVariables(), errors.Join(errors.New("task wait:"), err)
	}
	status := <-status_chan
	if status.ExitCode() != 0 {
		return cacao.NewVariables(), status.Error()
	}
	if err := client.Close(); err != nil {
		return cacao.NewVariables(), err
	}
	return readVariables(stdout.Bytes())
}
func ExecuteRemote(decoded []byte, metadata execution.Metadata, context capability.Context) (cacao.Variables, error) {
	filename := path.Join("lab", newRandomFilename())
	sshSession, sshClient, err := newSSHClient(context)
	if err != nil {
		log.Error("Could not establish ssh client")
		return cacao.NewVariables(), err
	}
	sftpClient, err := SSHCopyFile(decoded, sshClient, filename)
	if err != nil {
		log.Error("Could not copy file over ssh")
		return cacao.NewVariables(), err
	}
	if err := ExecuteRemoteFile(filename, sshSession); err != nil {
		log.Error("Could not execute remote file")
		return cacao.NewVariables(), err
	}
	variables, err := retrieveRemoteFile(filename, sftpClient)
	if err != nil {
		log.Error("Could not retrieve remote file")
		return cacao.NewVariables(), err
	}
	if err := deleteRemoteFile(filename, sftpClient); err != nil {
		log.Error("Could not delete remote file")
		return cacao.NewVariables(), err
	}
	if err := sftpClient.Close(); err != nil {
		log.Error("Could not close sftp client")
		return cacao.NewVariables(), err
	}
	if err := sshClient.Close(); err != nil {
		log.Error("Could not close ssh client")
		return cacao.NewVariables(), err
	}
	return variables, nil
}
func ExecuteRemoteFile(filename string, session *ssh.Session) error {
	cmd := path.Join("lab", "bin", "jupyter") + " nbconvert --execute --to notebook --inplace " + filename
	stdErr := strings.Builder{}
	session.Stderr = &stdErr
	output, err := session.Output(cmd)
	if err != nil {
		log.Trace("jupyter notebook ran unsuccessfully with output ", output)
		log.Trace("jupyter notebook ran unsuccessfully with stderr ", stdErr.String())
		return err
	}
	log.Trace("jupyter notebook ran successfully with output ", output)
	return nil
}
func retrieveRemoteFile(filename string, client *sftp.Client) (cacao.Variables, error) {
	file, err := client.Open(filename)
	if err != nil {
		return cacao.NewVariables(), err
	}
	stat, err := file.Stat()
	if err != nil {
		return cacao.NewVariables(), err
	}
	outputBuffer := make([]byte, stat.Size())
	if _, err := file.Read(outputBuffer); err != nil {
		return cacao.NewVariables(), err
	}
	if err := file.Close(); err != nil {
		return cacao.NewVariables(), nil
	}
	return readVariables(outputBuffer)
}
func deleteRemoteFile(filename string, client *sftp.Client) error {
	return client.Remove(filename)
}

type JupyterFile struct {
	Cells []JupyterCell `json:"cells" validate:"required"`
}
type JupyterCell struct {
	Type    string          `json:"cell_type" validate:"required"`
	Outputs []JupyterOutput `json:"outputs"`
}
type JupyterOutput struct {
	Name string   `json:"name" validate:"required"`
	Type string   `json:"output_type" validate:"required"`
	Text []string `json:"text" validate:"required"`
}

func readVariables(file_bytes []byte) (cacao.Variables, error) {
	var file JupyterFile
	if err := json.Unmarshal(file_bytes, &file); err != nil {
		return cacao.NewVariables(), err
	}
	variables := cacao.NewVariables()
	for _, cell := range file.Cells {
		for _, output := range cell.Outputs {
			for _, line := range output.Text {
				variable := readLine(line)
				if variable != nil {
					variables[variable.Name] = *variable
				}
			}
		}
	}
	log.Trace("Read variables ", variables)
	return variables, nil
}
func readLine(line string) *cacao.Variable {
	if cleaned_line, ok := strings.CutPrefix(line, "soarca::"); ok {
		cleaned_line = strings.TrimSpace(cleaned_line)
		// cleaned_line = strings.TrimSuffix(cleaned_line, "\n")
		line_content := strings.Split(cleaned_line, "=")
		if len(line_content) != 2 {
			log.Info("invalid variable assignment", cleaned_line)
			return nil
		}
		log.Infof("jupyter notebook resulted in assignment: %s = %s", line_content[0], line_content[1])
		return &cacao.Variable{
			Type:  "string",
			Name:  line_content[0],
			Value: line_content[1],
		}
	}
	return nil
}
func (capability *JupyterCapability) Execute(metadata execution.Metadata,
	context capability.Context) (cacao.Variables, error) {
	log.Trace("starting jupyter command execution")
	command := context.Command
	notebook_encoded := command.CommandB64
	if notebook_encoded == "" {
		return cacao.NewVariables(), errors.New("no notebook given in jupyter command")
	}
	return capability.ExecuteB64(notebook_encoded, metadata, context)
}
func (capability *JupyterCapability) GetType() string {
	return "jupyter"
}
func SSHCopyFile(file []byte, client *ssh.Client, destPath string) (*sftp.Client, error) {
	// open an SFTP session over an existing ssh connection.
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}

	// Create the destination file
	dstFile, err := sftpClient.Create(destPath)
	if err != nil {
		return nil, err
	}
	defer dstFile.Close()
	// write to file
	if _, err := dstFile.Write(file); err != nil {
		return nil, err
	}
	if err := dstFile.Close(); err != nil {
		return nil, err
	}
	return sftpClient, nil
}
