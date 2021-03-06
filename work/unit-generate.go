package work

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/blablacar/dgr/bin-templater/template"
	"github.com/blablacar/ggn/utils"
	"github.com/n0rad/go-erlog/logs"
	"github.com/peterbourgon/mergemap"
)

func (u *Unit) Generate(tmpl *template.Templating) error {
	u.generatedMutex.Lock()
	defer u.generatedMutex.Unlock()

	if u.generated {
		return nil
	}

	logs.WithFields(u.Fields).Debug("Generate")
	data := u.GenerateAttributes()
	aciList, err := u.Service.PrepareAcis()
	if err != nil {
		return err
	}
	acis := ""
	for _, aci := range aciList {
		acis += aci + " "
	}
	data["aciList"] = aciList
	data["acis"] = acis
	data["service_nodes"] = u.Service.nodesAsJsonMap

	out, err := json.Marshal(data)
	if err != nil {
		logs.WithEF(err, u.Fields).Panic("Cannot marshall attributes")
	}
	res := strings.Replace(string(out), "\\\"", "\\\\\\\"", -1)
	res = strings.Replace(res, "'", `\'`, -1)
	data["attributes"] = res
	data["attributesBase64"] = "base64," + base64.StdEncoding.EncodeToString([]byte(out))

	data["environmentAttributes"], data["environmentAttributesVars"] = u.prepareEnvironmentAttributes(data["attributes"].(string), "ATTR_")
	data["environmentAttributesBase64"], data["environmentAttributesVarsBase64"] = u.prepareEnvironmentAttributes(data["attributesBase64"].(string), "ATTR_BASE64_")

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		logs.WithEF(err, u.Fields).Error("Failed to run templating")
	}
	ok, err := utils.Exists(u.path)
	if !ok || err != nil {
		os.Mkdir(u.path, 0755)
	}
	err = ioutil.WriteFile(u.path+"/"+u.Filename, b.Bytes(), 0644)
	if err != nil {
		logs.WithEF(err, u.Fields).WithField("path", u.path+"/"+u.Filename).Error("Cannot writer unit")
	}

	u.generated = true
	return nil
}

func (u Unit) GenerateAttributes() map[string]interface{} {
	data := utils.CopyMap(u.Service.GetAttributes())
	data = mergemap.Merge(data, u.Service.NodeAttributes(u.hostname))
	return data
}

func (u Unit) prepareEnvironmentAttributes(data string, attrName string) (string, string) {
	var envAttr bytes.Buffer
	var envAttrVars bytes.Buffer
	var forbiddenSplitChar = []string{`:`, `.`, `"`, `,`, `'`, `*`, `=`, `\`}
	var shouldSplit bool

	num := 0
	for i, c := range data {
		y := i
		charBuffer := string(c)

		if i%1950 == 0 {
			shouldSplit = true
		}
		for y > 1 && (stringInSlice(string(data[y]), forbiddenSplitChar) || stringInSlice(string(data[y-1]), forbiddenSplitChar)) && shouldSplit {
			envAttr.Truncate(envAttr.Len() - 1)
			charBuffer = string(data[y-1]) + charBuffer
			y--
		}
		if shouldSplit {
			if num != 0 {
				envAttr.WriteString("'\n")
			}
			attrIndex := strconv.Itoa(num)
			envAttr.WriteString("Environment='" + attrName)
			envAttr.WriteString(attrIndex)
			envAttr.WriteString("=")
			envAttrVars.WriteString("${" + attrName)
			envAttrVars.WriteString(attrIndex)
			envAttrVars.WriteString("}")
			shouldSplit = false
			num++
		}
		envAttr.WriteString(charBuffer)
	}
	envAttr.WriteString("'\n")
	return envAttr.String(), envAttrVars.String()
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
