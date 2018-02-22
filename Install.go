package main

import (
	"infinidat-k8s-installer/lib"
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"github.com/howeyc/gopass"
)

// Apis supported by targeted kubernetes cluster
var supportedApisMap map[string]struct{}

//To take the flag "operation" while running the installer by default it will 'install'
var operationPtr = flag.String("operation", "install", "Specify the operation to perform (install/uninstall/storageclass) if not specified it will consider install")

//var skipOperation = flag.String("skip","","Specify operation which you want to skip eg prompt")
var installationPath = flag.String("path", "installation.yaml", "user can specify installation yaml")
var imagePullPath = flag.String("imagepath", "docker.io/infinidat/infinidat-k8s-provisioner:latest", "path to an image in a private repository ")
var imagePullsecret = flag.String("imagepullsecret", "", "secret in which credential of private repo are stored")
var loglevel = flag.Int("loglevel", 0, "to set the log level 0/1 1 for debug logs")

//incase of install and uninstall other wise it is getting updated
var input = "input.yaml"
var skipOperation string

func main() {
	flag.Parse()
	if *operationPtr != "install" && *operationPtr != "uninstall" && *operationPtr != "storageclass" {
		fmt.Println("Operation can be install, uninstall or storageclass")
		return
	}
	skipOperation = ""
	if strings.ToLower(flag.Arg(0)) == "noprompt" {
		skipOperation = skipOperation + "prompt"
	}

	if *operationPtr == "uninstall" { //incase of uninstall skip prompting
		skipOperation = skipOperation + "prompt"
	}
	if *operationPtr == "install" || *operationPtr == "uninstall" {
		//To check kubernetes is working or not
		isrunning, err := validateKubernetes()
		if err != nil {
			fmt.Println("kubernetes cluster not running ", err)
			return
		}
		if !isrunning {
			fmt.Println("kubernetes cluster not running ", err)
			return
		}
	}

	supportedApisMap = getSupportedApis()
	if *operationPtr == "storageclass" {
		value := "Which protocol will be supported by this storage class?\n 1. NFS \n 2. ISCSI \n 3. FC"
		var inputval string
		for {
			fmt.Println(value)
			fmt.Scanln(&inputval)
			if len(inputval) == 0 {
				start := strings.LastIndex(value, "(")
				end := strings.LastIndex(value, ")")
				if start == -1 || end == -1 {
					fmt.Println("default value not supported")
					continue
				}
				inputval = value[start+1 : end]
			}
			intInput, err := strconv.Atoi(inputval)
			if err != nil {
				fmt.Println("Please provide valid inputval ", err)
				continue
			}
			if intInput < 1 || intInput > 3 {
				fmt.Println("Input should be in range of 1 to 3", err)
				continue
			}
			break
		}
		fileappend := fmt.Sprint(time.Now().Format("20060102150405"))
		switch inputval {
		case "1":
			input = "input-nfs.yaml"
			*installationPath = "nfs-storageclass" + fileappend + ".yaml"
		case "2":
			input = "input-iscsi.yaml"
			*installationPath = "iscsi-storageclass" + fileappend + ".yaml"
		case "3":
			input = "input-fc.yaml"
			*installationPath = "fc-storageclass" + fileappend + ".yaml"
		}
	}

	//even thought prompt is skipped required yaml should exist
	if !(strings.Contains(skipOperation, "prompt") && outputExist()) {
		paramsTobePrompted := scan()
		valuesTobeReplaced := promptAndGetValues(paramsTobePrompted)
		err := replaceAndWriteYaml(valuesTobeReplaced)
		if err != nil {
			fmt.Errorf(err.Error())
			return
		}
	}

	switch *operationPtr {
	case "uninstall":
		err := uninstall()
		if err != nil {
			fmt.Println("uninstall failled", err)
		}
	case "install":
		err := install()
		if err != nil {
			fmt.Println("install failled", err)
		}
	case "storageclass":
		fmt.Println("Storage class yaml ", *installationPath, "  is created in current directory.")

	}
}
/*
returns boolean value
wheather file exist on installation path
 */
func outputExist() bool {
	if *operationPtr == "install" || *operationPtr == "uninstall" {
		_, err := os.Stat(*installationPath)
		return err == nil
	}
	return false
}
/*
return all supported apiversions
 */
func getSupportedApis() map[string]struct{} {
	cmd := exec.Command("kubectl", "api-versions")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmdError := &bytes.Buffer{}
	cmd.Stderr = cmdError
	err := cmd.Run()
	if err != nil {
		//return err
	}
	supportedApiList := strings.Split(cmdOutput.String(), "\n")
	supportedApisMap := make(map[string]struct{}, len(supportedApiList))
	for _, s := range supportedApiList {
		s = strings.Trim(s, " ")
		if len(s) > 0 {
			supportedApisMap[s] = struct{}{}
		}
	}
	return supportedApisMap
}
/*
runs kubectl delete on installation path
 */
func uninstall() (err error) {
	cmd := exec.Command("kubectl", "delete", "-f", *installationPath)
	cmd.Stdin = strings.NewReader("")
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmdErr := &bytes.Buffer{}
	cmd.Stderr = cmdErr
	fmt.Println(cmd.Args)
	err = cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprint(cmdErr))
	}
	fmt.Println(cmdOutput)
	return nil
}
/*
runs kubectl create on installation path
 */
func install() (err error) {
	cmd := exec.Command("kubectl", "create", "-f", *installationPath)
	cmd.Stdin = strings.NewReader("")
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmdErr := &bytes.Buffer{}
	cmd.Stderr = cmdErr
	fmt.Println(cmd.Args)
	err = cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprint(cmdErr))
	}
	fmt.Println(cmdOutput)
	return nil
}
/*
To check kubernetes is runing or not
 */
func validateKubernetes() (bool, error) {
	cmdcluster := exec.Command("kubectl", "cluster-info")
	cmdclusterOutput := &bytes.Buffer{}
	cmdcluster.Stdout = cmdclusterOutput
	err := cmdcluster.Run()

	if err != nil {
		return false, err
	}

	if !strings.Contains(fmt.Sprint(cmdclusterOutput), "is running") {
		return false, err
	}
	return true, nil

}
/*
prompt user for value
 */
func prompt(key,value string) string {
	for {
		var input string
		fmt.Println(value)
		if strings.Contains(strings.ToUpper(key),"PASSWORD"){
			barray,err:=gopass.GetPasswdMasked()
			if err==nil{
				input = string(barray)
			}
		}else{
			fmt.Scanln(&input)
		}
		if len(input) == 0 {
			start := strings.LastIndex(value, "(")
			end := strings.LastIndex(value, ")")
			if start == -1 || end == -1 {
				fmt.Println("default value not supported")
				continue
			}
			input = value[start+1 : end]
		}
		return input
	}
}

func promptAndGetValues(paramsTobePrompted *lib.LinkedMap) map[string]string {
	copyOfParams := make(map[string]string)
	iterator := paramsTobePrompted.GetIterator()
	for k, v := iterator(); k != nil; k, v = iterator() {
		if strings.HasPrefix(*k, "#$") {
			copyOfParams[*k] = prompt(*k,*v)
		}
		if strings.HasPrefix(*k, "apiVersion") {
			//will not prompt
			//checks wheather api version is supported by kubernetes enviorment or not
			//if not will find appropriated supported api version
			index := strings.Index(*k, ":")
			key := *k
			apiVer := strings.Trim(key[index+1:], " ")

			if _, ok := supportedApisMap[apiVer]; ok {
				copyOfParams[*k] = *k
			} else { //finding appropriate supported version for given api
				if strings.Contains(apiVer, "/") {
					oldapi := strings.Split(apiVer, "/")
					expectedApi := oldapi[0]
					//finding expected api in supported apis with different versions
					for supportedApi := range supportedApisMap {
						supportedApi = strings.Trim(supportedApi, " ")
						if strings.HasPrefix(supportedApi, expectedApi) {
							copyOfParams[*k] = "apiVersion: " + supportedApi
						}
					}
				}
			}
		}
	}
	//static params
	logstatic := "\"-v=" + fmt.Sprint(*loglevel) + "\""
	copyOfParams["$loglevel"] = logstatic
	if len(*imagePullsecret) > 0 {
		secret := "\n      imagePullSecrets: \n      - name: " + *imagePullsecret
		copyOfParams["$imagepullsecret"] = secret
	} else {
		copyOfParams["$imagepullsecret"] = " "
	}
	copyOfParams["$imagepath"] = *imagePullPath
	return copyOfParams
}
/*
prepares yaml files conidering user inputs
 */
func replaceAndWriteYaml(replace map[string]string) error {
	file, err := os.Create(*installationPath)
	if err != nil {
		return err
	}
	defer file.Close()

	inFile, _ := os.Open(input)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	w := bufio.NewWriter(file)
	for scanner.Scan() {
		currenyline := scanner.Text()
		if !strings.HasPrefix(currenyline, "#") {
			for k, v := range replace {
				k = strings.Trim(k, "#")
				if strings.Contains(k, "base64") {
					v = base64.StdEncoding.EncodeToString([]byte(v))
				}
				if strings.Contains(k, "Bool") {
					v = "\"" + v + "\""
				}

				currenyline = strings.Replace(currenyline, k, v, -1)
			}
			fmt.Fprintln(w, currenyline)
		}
	}
	return w.Flush()
}
/*
scans input yamls
prepares list for hich installer needs to prompt for value to user
 */
func scan() *lib.LinkedMap {
	promptingParams := lib.LinkedMap{}
	inFile, _ := os.Open(input)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		currenyline := strings.Trim(scanner.Text(), " ")

		if strings.HasPrefix(currenyline, "#$") && strings.Contains(currenyline, "=") {
			index := strings.Index(currenyline, "=")
			key := strings.Trim(currenyline[:index], " ")
			//key = strings.Trim(key, "#")
			value := strings.Trim(currenyline[index+1:], " ")
			promptingParams.Put(key, strings.Trim(value, " "))
		} else if strings.HasPrefix(currenyline, "apiVersion") && strings.Contains(currenyline, ":") {
			currenyline = strings.Trim(currenyline, " ")
			promptingParams.Put(currenyline, currenyline)
		}
	}

	return &promptingParams
}
