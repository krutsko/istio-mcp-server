package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/istio-mcp-server/pkg/mcp"
	"github.com/istio-mcp-server/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
)

var rootCmd = &cobra.Command{
	Use:   "istio-mcp-server [command] [options]",
	Short: "Istio Model Context Protocol (MCP) server",
	Long: `
Istio Model Context Protocol (MCP) server

A comprehensive MCP server for interacting with Istio service mesh resources.
Provides tools for querying Virtual Services, Destination Rules, Gateways,
and Envoy proxy configurations.

  # show this help
  istio-mcp-server -h

  # shows version information
  istio-mcp-server --version

  # start STDIO server
  istio-mcp-server

  # start a SSE server on port 8080
  istio-mcp-server --sse-port 8080

  # start a SSE server on port 8443 with a public HTTPS host of example.com
  istio-mcp-server --sse-port 8443 --sse-base-url https://example.com:8443

  # start with custom kubeconfig and read-only mode
  istio-mcp-server --kubeconfig ~/.kube/config --read-only

  # start HTTP server on port 8080`,
	Run: func(cmd *cobra.Command, args []string) {
		initLogging()
		profile := mcp.ProfileFromString(viper.GetString("profile"))
		if profile == nil {
			fmt.Printf("Invalid profile name: %s, valid names are: %s\n", viper.GetString("profile"), strings.Join(mcp.ProfileNames, ", "))
			os.Exit(1)
		}

		klog.V(1).Info("Starting istio-mcp-server")
		klog.V(1).Infof(" - Profile: %s", profile.GetName())

		if viper.GetBool("version") {
			fmt.Println(version.Version)
			return
		}
		mcpServer, err := mcp.NewServer(mcp.Configuration{
			Profile:    profile,
			Kubeconfig: viper.GetString("kubeconfig"),
		})
		if err != nil {
			fmt.Printf("Failed to initialize MCP server: %v\n", err)
			os.Exit(1)
		}
		defer mcpServer.Close()

		ssePort := viper.GetInt("sse-port")
		if ssePort > 0 {
			sseServer := mcpServer.ServeSse(viper.GetString("sse-base-url"))
			defer func() { _ = sseServer.Shutdown(cmd.Context()) }()
			klog.V(0).Infof("SSE server starting on port %d and path /sse", ssePort)
			if err := sseServer.Start(fmt.Sprintf(":%d", ssePort)); err != nil {
				klog.Errorf("Failed to start SSE server: %s", err)
				return
			}
		}

		httpPort := viper.GetInt("http-port")
		if httpPort > 0 {
			httpServer := mcpServer.ServeHTTP()
			klog.V(0).Infof("Streaming HTTP server starting on port %d and path /mcp", httpPort)
			if err := httpServer.Start(fmt.Sprintf(":%d", httpPort)); err != nil {
				klog.Errorf("Failed to start streaming HTTP server: %s", err)
				return
			}
		}

		if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		klog.Errorf("Failed to execute command: %s", err)
		os.Exit(1)
	}
}

// initLogging initializes the logging configuration
func initLogging() {
	flagSet := flag.NewFlagSet("istio-mcp-server", flag.ContinueOnError)
	klog.InitFlags(flagSet)
	loggerOptions := []textlogger.ConfigOption{textlogger.Output(os.Stdout)}
	if logLevel := viper.GetInt("log-level"); logLevel >= 0 {
		loggerOptions = append(loggerOptions, textlogger.Verbosity(logLevel))
		_ = flagSet.Parse([]string{"--v", strconv.Itoa(logLevel)})
	}
	logger := textlogger.NewLogger(textlogger.NewConfig(loggerOptions...))
	klog.SetLoggerWithOptions(logger)
}

// flagInit initializes the flags for the root command.
// Exposed for testing purposes.
func flagInit() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	rootCmd.Flags().IntP("log-level", "", 0, "Set the log level (from 0 to 9)")
	rootCmd.Flags().IntP("sse-port", "", 0, "Start a SSE server on the specified port")
	rootCmd.Flags().IntP("http-port", "", 0, "Start a streamable HTTP server on the specified port")
	rootCmd.Flags().StringP("sse-base-url", "", "", "SSE public base URL to use when sending the endpoint message (e.g. https://example.com)")
	rootCmd.Flags().StringP("kubeconfig", "", "", "Path to the kubeconfig file to use for authentication")
	rootCmd.Flags().String("profile", "full", "MCP profile to use (one of: "+strings.Join(mcp.ProfileNames, ", ")+")")

	_ = viper.BindPFlags(rootCmd.Flags())
}

func init() {
	flagInit()
}
