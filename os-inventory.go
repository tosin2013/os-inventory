package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// New loads default values and returns *Inventory
func New(defaultValues map[string]interface{}) *Inventory {
	inv := &Inventory{
		GeneratorDeploymentType:        defaultValues["generatorDeploymentType"].(string),
		GeneratorSshUser:               defaultValues["generatorSshUser"].(string),
		GeneratorNfsEnabled:            defaultValues["generatorNfsEnabled"].(bool),
		GeneratorRegistryNativeNfs:     defaultValues["generatorRegistryNativeNfs"].(bool),
		GeneratorHaproxyEnabled:        defaultValues["generatorHaproxyEnabled"].(bool),
		GeneratorInstallVersion:        defaultValues["generatorInstallVersion"].(string),
		GeneratorSkipChecks:            defaultValues["generatorSkipChecks"].(bool),
		GeneratorMultiMaster:           defaultValues["generatorMultiMaster"].(bool),
		GeneratorClusterMethod:         defaultValues["generatorClusterMethod"].(string),
		GeneratorClusterHostname:       defaultValues["generatorClusterHostname"].(string),
		GeneratorClusterPublicHostname: defaultValues["generatorClusterPublicHostname"].(string),
		GeneratorContainerizedDeploy:   defaultValues["generatorContainerizedDeploy"].(bool),
		GeneratorContainerizedOvs:      defaultValues["generatorContainerizedOvs"].(bool),
		GeneratorContainerizedNode:     defaultValues["generatorContainerizedNode"].(bool),
		GeneratorContainerizedMaster:   defaultValues["generatorContainerizedMaster"].(bool),
		GeneratorContainerizedEtcd:     defaultValues["generatorContainerizedEtcd"].(bool),
		GeneratorSystemImagesRegistry:  defaultValues["generatorSystemImagesRegistry"].(string),
		GeneratorOpenshiftUseCrio:      defaultValues["generatorOpenshiftUseCrio"].(bool),
		GeneratorOpenshiftCrioUseRpm:   defaultValues["generatorOpenshiftCrioUseRpm"].(bool),
		GeneratorMultiInfra:            defaultValues["generatorMultiInfra"].(bool),
		GeneratorUseXip:                defaultValues["generatorUseXip"].(bool),
		GeneratorInfraIpv4:             defaultValues["generatorInfraIpv4"].(string),
		GeneratorExtDnsWildcard:        defaultValues["generatorExtDnsWildcard"].(string),
		GeneratorSdnPlugin:             defaultValues["generatorSdnPlugin"].(string),
		GeneratorDisableServiceCatalog: defaultValues["generatorDisableServiceCatalog"].(bool),
		GeneratorInfraReplicas:         defaultValues["generatorInfraReplicas"].(int),
		GeneratorMetricsEnabled:        defaultValues["generatorMetricsEnabled"].(bool),
		GeneratorDeployHosa:            defaultValues["generatorDeployHosa"].(bool),
		GeneratorMetricsNativeNfs:      defaultValues["generatorMetricsNativeNfs"].(bool),
		GeneratorPrometheusEnabled:     defaultValues["generatorPrometheusEnabled"].(bool),
		GeneratorPrometheusNativeNfs:   defaultValues["generatorPrometheusNativeNfs"].(bool),
		GeneratorLoggingEnabled:        defaultValues["generatorLoggingEnabled"].(bool),
		GeneratorLoggingNativeNfs:      defaultValues["generatorLoggingNativeNfs"].(bool),
		GeneratorMastersList:           defaultValues["generatorMastersList"].([]string),
		GeneratorEtcdList:              defaultValues["generatorEtcdList"].([]string),
		GeneratorLbList:                defaultValues["generatorLbList"].([]string),
		GeneratorNodesMap:              defaultValues["generatorNodesMap"].(map[string]string),
	}
	return inv
}

// ParseYAML parses a YAML input file for custom paramters values
func parseYAML(yamlFile string, inv *Inventory) error {
	data, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &inv)
	if err != nil {
		return err
	}
	return nil
}

// doSanityChecks verifies passed parameters
func doSanityChecks(inv *Inventory) error {

	err := inv.CheckDeploymentType()
	if err != nil {
		return err
	}
	err = inv.CheckInstallVersion()
	if err != nil {
		return err
	}
	err = inv.CheckClusterMethod()
	if err != nil {
		return err
	}
	err = inv.CheckInfraIpv4()
	if err != nil {
		return err
	}
	err = inv.CheckSdnPlugin()
	if err != nil {
		return err
	}

	return nil
}

const appDescription = `OpenShift installation inventory generation tool.
	 The OpenShift adavanced installation method uses Ansible to provide a
	 flexible and reliable way to deploy OpenShift on enterprise grade clusters.
	 The whole deployments relies on a rich Ansible inventory where all nodes
	 are defined, along with a huge set of inventory variables.
	 Sometimes creating a basic inventory ready to be customized can be a long
	 process.
	 The purpose of os-inventory is to ease the inventory creation process, yet
	 leaving to the user the freedom to apply further customizations.`

func main() {

	var loadYAML string
	var inventoryFile string
	var defaultsFile string

	inventory := New(defaults)

	app := cli.NewApp()
	app.Name = "os-inventory"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Giovan Battista Salinetti",
			Email: "gbsalinetti@gmail.com",
		},
	}
	app.Usage = "OpenShift installation inventory generation tool"
	app.Description = appDescription
	app.Version = "0.1.3"
	app.Commands = []cli.Command{
		{
			Name:        "generate",
			Aliases:     []string{"gen", "g"},
			Usage:       "Generates the inventory file for OpenShift installations",
			Description: "Generates the inventory file for OpenShift installations",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "file, f",
					Usage:       "Load a YAML configuration file",
					Destination: &loadYAML,
				},
				cli.StringFlag{
					Name:        "output, o",
					Usage:       "Print generated inventory to file",
					Destination: &inventoryFile,
				},
			},
			Action: func(c *cli.Context) error {
				// Load YAML configuration if passed
				if loadYAML != "" {
					filePath := loadYAML
					err := parseYAML(filePath, inventory)
					if err != nil {
						return err
					}
				}
				// Create new template and parse content
				t, err := template.New("OpenShiftInventory").Parse(tmpl)
				if err != nil {
					return err
				}
				// Run sanity checks before exporting
				err = doSanityChecks(inventory)
				if err != nil {
					return err
				}
				// Generate the processed inventory
				if inventoryFile != "" {
					f, err := os.Create(inventoryFile)
					if err != nil {
						return err
					}
					// Print inventory to file
					err = t.Execute(f, inventory)
					if err != nil {
						return err
					}
				} else {
					// Print inventory to stdout
					err = t.Execute(os.Stdout, inventory)
					if err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			Name:        "defaults",
			Aliases:     []string{"def", "d"},
			Usage:       "Prints default configuration in YAML format",
			Description: "Prints default configuration in YAML format",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "output, o",
					Usage:       "Print defaults to file",
					Destination: &defaultsFile,
				},
			},
			Action: func(c *cli.Context) error {
				d, err := yaml.Marshal(&inventory)
				if err != nil {
					return err
				}
				if defaultsFile != "" {
					f, err := os.Create(defaultsFile)
					if err != nil {
						return err
					}
					fmt.Fprintf(f, "---\n%s", d)
				} else {
					fmt.Printf("---\n%s", d)
				}
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
