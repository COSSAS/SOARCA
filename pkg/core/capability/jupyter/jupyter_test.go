package jupyter

import (
	"soarca/pkg/models/cacao"
	"testing"

	"github.com/stretchr/testify/assert"
)

var converted_playbook = `{
	"cells": [
  {
   "cell_type": "code",
   "execution_count": 1,
   "id": "7596342c",
   "metadata": {
    "execution": {
     "iopub.execute_input": "2026-02-09T11:03:21.419544Z",
     "iopub.status.busy": "2026-02-09T11:03:21.419382Z",
     "iopub.status.idle": "2026-02-09T11:03:21.423758Z",
     "shell.execute_reply": "2026-02-09T11:03:21.422894Z"
    }
   },
   "outputs": [
    {
     "name": "stdout",
     "output_type": "stream",
     "text": [
      "soarca::__hello__=world\n"
     ]
    }
   ],
   "source": [
    "print(\"soarca::__hello__=world\")"
   ]
  }
 ],
 "metadata": {
  "language_info": {
   "codemirror_mode": {
    "name": "ipython",
    "version": 3
   },
   "file_extension": ".py",
   "mimetype": "text/x-python",
   "name": "python",
   "nbconvert_exporter": "python",
   "pygments_lexer": "ipython3",
   "version": "3.14.2"
  }
 },
 "nbformat": 4,
 "nbformat_minor": 5
}`

func TestReadLine(t *testing.T) {
	line1 := "soarca::__hello__=world"
	variable := readLine(line1)
	assert.NotNil(t, variable)
	assert.Equal(t, variable.Name, "__hello__")
	assert.Equal(t, variable.Value, "world")
}

func TestReadVariables(t *testing.T) {
	variables, err := readVariables([]byte(converted_playbook))
	assert.Nil(t, err)
	assert.Len(t, variables, 1)
	variable, ok := variables["__hello__"]
	assert.True(t, ok)
	assert.NotNil(t, variable)
	assert.Equal(t, variable.Name, "__hello__")
	assert.Equal(t, variable.Value, "world")
}

func TestGetUrl(t *testing.T) {
	ipv4_addr := cacao.Addresses{"ipv4": []string{"1.2.3.4"}}
	port := "123"
	url := getUrl(ipv4_addr, port)
	assert.Equal(t, url, "1.2.3.4:123")
	url_addr := cacao.Addresses{"url": []string{"test.com/jupyter"}}
	url = getUrl(url_addr, port)
	// No port should be set in the url
	assert.Equal(t, url, "test.com/jupyter")
}
