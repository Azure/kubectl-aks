// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"fmt"
	"os"
	"sort"

	"github.com/Azure/kubectl-aks/cmd/utils/config"
	"github.com/manifoldco/promptui"
)

const allNodesLabel = "(all nodes in cluster)"

// SelectionResult holds the user's interactive selection.
type SelectionResult struct {
	Cluster  string
	Node     string   // empty when AllNodes is true
	AllNodes bool     // true if user chose to target entire cluster
	Nodes    []string // populated when AllNodes is true
}

// InteractiveSelectNode prompts the user to pick a cluster and node.
// If a current cluster is set, it uses that cluster directly and only
// prompts for the node. Otherwise it prompts for both.
// Returns an error if stdin is not a terminal.
func InteractiveSelectNode(cfg *config.Config) (*SelectionResult, error) {
	if !isTerminal() {
		return nil, fmt.Errorf("interactive selection requires a terminal (TTY)")
	}

	clusters, err := cfg.ListClusters()
	if err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}
	if len(clusters) == 0 {
		return nil, fmt.Errorf("no clusters configured; run 'kubectl-aks config import' first")
	}
	sort.Strings(clusters)

	// Use current-cluster if set; otherwise prompt.
	clusterName := cfg.CurrentClusterName()
	if clusterName == "" {
		if len(clusters) == 1 {
			clusterName = clusters[0]
		} else {
			prompt := promptui.Select{
				Label: "Select cluster",
				Items: clusters,
			}
			_, clusterName, err = prompt.Run()
			if err != nil {
				return nil, fmt.Errorf("cluster selection: %w", err)
			}
		}
	}

	nodes, err := cfg.ListClusterNodes(clusterName)
	if err != nil {
		return nil, fmt.Errorf("listing nodes for cluster %s: %w", clusterName, err)
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("cluster %q has no nodes", clusterName)
	}
	sort.Strings(nodes)

	items := append([]string{allNodesLabel}, nodes...)
	prompt := promptui.Select{
		Label: fmt.Sprintf("Select node in %s", clusterName),
		Items: items,
	}
	_, selected, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("node selection: %w", err)
	}

	if selected == allNodesLabel {
		return &SelectionResult{
			Cluster:  clusterName,
			AllNodes: true,
			Nodes:    nodes,
		}, nil
	}

	return &SelectionResult{
		Cluster: clusterName,
		Node:    selected,
	}, nil
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
