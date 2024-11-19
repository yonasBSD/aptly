package api

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"os"
	"os/exec"

	"github.com/aptly-dev/aptly/deb"
	"github.com/gin-gonic/gin"
)

// @Summary Graph Output
// @Description **Generate graph of aptly objects (same as in aptly graph command).**
// @Description e.g open url `http://localhost:8080/api/graph.svg?layout=vertical` in browser
// @Tags Graph
// @Produce json
// @Param ext path string true "ext specifies desired file extension, e.g. .png, .svg."
// @Param layout query string false "Change between a `horizontal` (default) and a `vertical` graph layout."
// @Success 200 {object} []byte "Output"
// @Failure 404 {object} Error "Not Found"
// @Failure 500 {object} Error "Internal Server Error"
// @Router /api/graph [get]
// GET /api/graph.:ext?layout=[vertical|horizontal(default)]
func apiGraph(c *gin.Context) {
	var (
		err    error
		output []byte
	)

	ext := c.Params.ByName("ext")
	layout := c.Request.URL.Query().Get("layout")
	factory := context.NewCollectionFactory()

	graph, err := deb.BuildGraph(factory, layout)
	if err != nil {
		c.JSON(500, err)
		return
	}

	buf := bytes.NewBufferString(graph.String())

	if ext == "dot" || ext == "gv" {
		// If the raw dot data is requested, return it as string.
		// This allows client-side rendering rather than server-side.
		c.String(200, buf.String())
		return
	}

	command := exec.Command("dot", "-T"+ext)
	command.Stderr = os.Stderr

	stdin, err := command.StdinPipe()
	if err != nil {
		AbortWithJSONError(c, 500, err)
		return
	}

	_, err = io.Copy(stdin, buf)
	if err != nil {
		AbortWithJSONError(c, 500, err)
		return
	}

	err = stdin.Close()
	if err != nil {
		AbortWithJSONError(c, 500, err)
		return
	}

	output, err = command.Output()
	if err != nil {
		AbortWithJSONError(c, 500, fmt.Errorf("unable to execute dot: %s (is graphviz package installed?)", err))
		return
	}

	mimeType := mime.TypeByExtension("." + ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Data(200, mimeType, output)
}
