package main

import (
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"strconv"
	"time"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Connection
	VCenterHost string `mapstructure:"vcenter_host"`
	Datacenter  string `mapstructure:"datacenter"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`

	// Location
	Template     string `mapstructure:"template"`
	FolderName   string `mapstructure:"folder"`
	VMName       string `mapstructure:"vm_name"`
	Host         string `mapstructure:"host"`
	ResourcePool string `mapstructure:"resource_pool"`
	Datastore    string `mapstructure:"datastore"`
	LinkedClone  bool   `mapstructure:"linked_clone"`

	// Customization
	CPUs string `mapstructure:"CPUs"`
	RAM  string `mapstructure:"RAM"`

	// Provisioning
	communicator.Config `mapstructure:",squash"`

	// Post-processing
	ShutdownCommand    string `mapstructure:"shutdown_command"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`
	ShutdownTimeout    time.Duration
	CreateSnapshot     bool   `mapstructure:"create_snapshot"`
	ConvertToTemplate  bool   `mapstructure:"convert_to_template"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	{
		err := config.Decode(c, &config.DecodeOpts{
			Interpolate:        true,
			InterpolateContext: &c.ctx,
		}, raws...)
		if err != nil {
			return nil, nil, err
		}
	}

	// Accumulate any errors
	errs := new(packer.MultiError)
	var warnings []string

	// Prepare config(s)
	errs = packer.MultiErrorAppend(errs, c.Config.Prepare(&c.ctx)...)

	// Check the required params
	if c.VCenterHost == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("vCenter host required"))
	}
	if c.Username == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Username required"))
	}
	if c.Password == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Password required"))
	}
	if c.Template == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Template VM name required"))
	}
	if c.VMName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Target VM name required"))
	}
	if c.Host == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Target host required"))
	}

	// Verify numeric parameters if present
	if c.CPUs != "" {
		if _, err := strconv.Atoi(c.CPUs); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Invalid number of CPU sockets"))
		}
	}
	if c.RAM != "" {
		if _, err := strconv.Atoi(c.RAM); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Invalid number for RAM"))
		}
	}
	if c.RawShutdownTimeout != "" {
		timeout, err := time.ParseDuration(c.RawShutdownTimeout)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
		}
		c.ShutdownTimeout = timeout
	} else {
		c.ShutdownTimeout = 5 * time.Minute
	}

	if len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}
