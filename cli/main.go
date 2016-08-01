package main

import (
	"fmt"
	"github.com/ghetzel/cli"
	"github.com/ghetzel/pivot"
	"github.com/ghetzel/pivot/filter"
	"github.com/ghetzel/pivot/filter/generators"
	"github.com/ghetzel/pivot/util"
	"github.com/op/go-logging"
	"os"
)

var log = logging.MustGetLogger(`main`)

func main() {
	app := cli.NewApp()
	app.Name = util.ApplicationName
	app.Usage = util.ApplicationSummary
	app.Version = util.ApplicationVersion
	app.EnableBashCompletion = false

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `info`,
			EnvVar: `LOGLEVEL`,
		},
	}

	app.Before = func(c *cli.Context) error {
		logging.SetFormatter(logging.MustStringFormatter(`%{color}%{level:.4s}%{color:reset}[%{id:04d}] %{message}`))

		if level, err := logging.LogLevel(c.String(`log-level`)); err == nil {
			logging.SetLevel(level, `main`)
		} else {
			return err
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  `serve`,
			Usage: `Start the HTTP data proxy service.`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  `config, c`,
					Usage: `Path to the configuration file to load.`,
					Value: `/etc/pivot.yml`,
				},
				cli.StringFlag{
					Name:  `address, a`,
					Usage: `The local address the server should listen on.`,
					Value: pivot.DEFAULT_SERVER_ADDRESS,
				},
				cli.IntFlag{
					Name:  `port, p`,
					Usage: `The port the server should listen on.`,
					Value: pivot.DEFAULT_SERVER_PORT,
				},
			},
			Action: func(c *cli.Context) {
				if config, err := pivot.LoadConfigFile(c.String(`config`)); err == nil {
					if err := config.Initialize(); err == nil {
						// start monitoring backends
						go pivot.MonitorBackends()

						server := pivot.NewServer()
						server.Address = c.String(`address`)
						server.Port = c.Int(`port`)

						if err := server.ListenAndServe(); err != nil {
							log.Fatalf("Failed to start server: %v", err)
							os.Exit(3)
						}
					} else {
						log.Fatalf("Failed to initialize: %v", err)
						os.Exit(2)
					}
				} else {
					log.Fatalf("Failed to load configuration: %v", err)
					os.Exit(1)
				}
			},
		},
		{
			Name:      `filter`,
			ArgsUsage: `TYPE COLLECTION SPEC`,
			Usage:     `Parse a filter specificiation and generate type-specific query syntax.`,
			Action: func(c *cli.Context) {
				if len(c.Args()) > 2 {
					generatorType := c.Args()[0]
					collection := c.Args()[1]
					spec := c.Args()[2]

					var generator filter.IGenerator

					switch generatorType {
					case `sql92`:
						generator = generators.NewSql92Generator()
					case `elasticsearch`:
						generator = generators.NewElasticsearchGenerator()
					default:
						log.Fatalf("Unknown generator type '%s'", generatorType)
					}

					if f, err := filter.Parse(spec); err == nil {
						if payload, err := filter.Render(generator, collection, f); err == nil {
							fmt.Printf("%s\n", payload)
						} else {
							log.Fatalf("Failed to render: %v", err)
						}
					} else {
						log.Fatalf("Failed to parse query: %v", err)
					}
				} else {
					log.Fatalf("Not enough arguments")
				}
			},
		},
	}

	app.Run(os.Args)
}
