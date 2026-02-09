package jupyter

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/utils"
	"strings"
)

type JupyterCapability struct {
}

var component = reflect.TypeOf(JupyterCapability{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Trace, "", logger.Json)
}

func New() JupyterCapability {
	return JupyterCapability{}
}

func getUrl(addresses cacao.Addresses, port string) string {
	if url, ok := addresses["url"]; ok {
		if len(url) > 0 {
			return url[0]
		}
	}
	if ipv4, ok := addresses["ipv4"]; ok {
		if len(ipv4) > 0 {
			if len(port) > 0 {
				return fmt.Sprintf("%s:%s", ipv4[0], port)

			}
			return ipv4[0]
		}
	}
	if ipv6, ok := addresses["ipv6"]; ok {
		if len(ipv6) > 0 {
			if len(port) > 0 {
				return fmt.Sprintf("%s:%s", ipv6[0], port)

			}
			return ipv6[0]
		}
	}
	return utils.GetEnv("JUPYTER_URI", "http://jupyter:7000")
}

func (capability *JupyterCapability) ExecuteB64(encoded string, addresses cacao.Addresses, port string) (cacao.Variables, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Error("Could not decode b64 notebook")
		return cacao.NewVariables(), err
	}

	body := bytes.NewReader(decoded)
	jupyter_addr := getUrl(addresses, port)
	response, err := http.Post(jupyter_addr, "", body)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	responseBytes, err := io.ReadAll(response.Body)
	if response.Body.Close() != nil {
		log.Warning("error closing response body")
	}
	if response.StatusCode != 200 {
		response_str := string(responseBytes)
		err = fmt.Errorf("Jupyter server returned %d\n%s", response.StatusCode, response_str)
		return cacao.NewVariables(), err
	}
	return readVariables(responseBytes)
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
	target_addr := context.Target.Address
	port := context.Target.Port
	if notebook_encoded == "" {
		return cacao.NewVariables(), errors.New("no notebook given in jupyter command")
	}
	return capability.ExecuteB64(notebook_encoded, target_addr, port)
}
func (capability *JupyterCapability) GetType() string {
	return "jupyter"
}
