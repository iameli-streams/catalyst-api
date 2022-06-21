package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	serfclient "github.com/hashicorp/serf/client"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/mitchellh/cli"
	"github.com/peterbourgon/ff/v3"
)

type (
	catalystConfig struct {
		serfRPCAddress           string
		serfRPCAuthKey           string
		mistLoadBalancerEndpoint string
	}
)

var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}

	// eli note: this is copied from here:
	// https://github.com/hashicorp/serf/blob/a2bba5676d6e37953715ea10e583843793a0c507/cmd/serf/commands.go#L20-L25
	// but we should someday get a little bit smarter and invoke serf directly
	// instead of wrapping their CLI helper

	Commands = map[string]cli.CommandFactory{
		"agent": func() (cli.Command, error) {
			a := &agent.Command{
				Ui:         ui,
				ShutdownCh: make(chan struct{}),
			}
			return a, nil
		},
	}
}

func runClient(config catalystConfig) error {
	// eli note: hardcoded. needs to parse out configuration from the CLI

	client, err := connectSerfAgent(config.serfRPCAddress, config.serfRPCAuthKey)

	if err != nil {
		return err
	}
	defer client.Close()

	eventCh := make(chan map[string]interface{}, 1024)
	streamHandle, err := client.Stream("*", eventCh)
	if err != nil {
		return fmt.Errorf("error starting stream: %s", err)
	}
	defer client.Stop(streamHandle)

	// eli note: not sure if this is useful yet, but we can do this as well:

	// logCh := make(chan string, 1024)
	// monHandle, err := client.Monitor(logutils.LogLevel("INFO"), logCh)
	// if err != nil {
	// 	return fmt.Errorf("error starting monitor: %s", err)
	// }
	// defer client.Stop(monHandle)

	// eli note: uncertain how we handle dis/reconnects here. but it's local, so hopefully rare?
	for {
		event := <-eventCh
		fmt.Printf(" got event: %v\n", event)

		members, err := getSerfMembers(client)

		if err != nil {
			return err
		}

		balancedServers, err := getMistLoadBalancerServers(config.mistLoadBalancerEndpoint)

		if err != nil {
			fmt.Printf("Error getting mist load balancer servers: %s\n", err)
			return err
		}

		membersMap := make(map[string]bool)

		for _, member := range members {
			member_host := member.Addr.String()

			// commented out as for now the load balancer does not return ports
			//if member.Port != 0 {
			//	member_host = fmt.Sprintf("%s:%d", member_host, member.Port)
			//}

			membersMap[member_host] = true
		}

		fmt.Printf("current members in cluster: %v\n", membersMap)
		fmt.Printf("current members in load balancer: %v\n", balancedServers)

		// compare membersMap and balancedServers
		// del all servers not present in membersMap but present in balancedServers
		// add all servers not present in balancedServers but present in membersMap

		// note: untested as per MistUtilLoad ports
		for k := range balancedServers {
			if _, ok := membersMap[k]; !ok {
				fmt.Printf("deleting server %s from load balancer\n", k)
				changeLoadBalancerServers(config.mistLoadBalancerEndpoint, k, "del")
			}
		}

		for k := range membersMap {
			if _, ok := balancedServers[k]; !ok {
				fmt.Printf("adding server %s to load balancer\n", k)
				changeLoadBalancerServers(config.mistLoadBalancerEndpoint, k, "add")
			}
		}

	}

	return nil
}

func connectSerfAgent(serfRPCAddress string, serfRPCAuthKey string) (*serfclient.RPCClient, error) {
	return serfclient.ClientFromConfig(&serfclient.Config{
		Addr:    serfRPCAddress,
		AuthKey: serfRPCAuthKey,
	})
}

func getSerfMembers(client *serfclient.RPCClient) ([]serfclient.Member, error) {
	return client.Members()
}

func changeLoadBalancerServers(endpoint string, server string, action string) ([]byte, error) {
	url := endpoint + "?" + action + "server=" + url.QueryEscape(server)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error response from load balancer changing servers: %s\n", string(b))
		return b, errors.New(string(b))
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s", err)
		return nil, err
	}
	fmt.Println("requested mist to " + action + " server " + server + " to the load balancer")
	fmt.Println(string(b))
	return b, nil
}

func getMistLoadBalancerServers(endpoint string) (map[string]interface{}, error) {

	url := endpoint + "?lstservers=1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error response from load balancer listing servers: %s\n", string(b))
		return nil, errors.New(string(b))
	}
	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Error reading response: %s", err)
		return nil, err
	}

	var mistResponse map[string]interface{}

	json.Unmarshal([]byte(string(b)), &mistResponse)

	return mistResponse, nil
}

func main() {
	args := os.Args[1:]

	flag.Set("logtostderr", "true")
	fs := flag.NewFlagSet("catalyst-node-connected", flag.ExitOnError)

	serfRPCAddress := fs.String("serf-rpc-address", "127.0.0.1:7373", "Serf RPC address")
	serfRPCAuthKey := fs.String("serf-rpc-auth-key", "", "Serf RPC auth key")
	mistLoadBalancerEndpoint := fs.String("mist-load-balancer-endpoint", "http://127.0.0.1:8042/", "Mist util load endpoint")

	ff.Parse(
		fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("CATALYST_NODE"),
		ff.WithEnvVarSplit(","),
	)
	flag.CommandLine.Parse(nil)

	config := catalystConfig{
		serfRPCAddress:           *serfRPCAddress,
		serfRPCAuthKey:           *serfRPCAuthKey,
		mistLoadBalancerEndpoint: *mistLoadBalancerEndpoint,
	}

	go func() {
		// eli note: i put this in a loop in case client boots before server.
		// doesn't seem to happen in practice.
		for {
			err := runClient(config)
			if err != nil {
				fmt.Printf("Error starting client: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	cli := &cli.CLI{
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.BasicHelpFunc("catalyst-node"),
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(exitCode)
}
