package core

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type DiffService struct {
}

// DiffType represents the type of difference found
type DiffType int

const (
	Added DiffType = iota
	Removed
	Modified
)

// Diff represents a single difference between two HTML documents
type Diff struct {
	Type    DiffType
	Path    string // XPath-like path to the node
	Old     string // Old value (empty for Added)
	New     string // New value (empty for Removed)
	Message string
}

func NewDiffService() *DiffService {
	diffService := new(DiffService)
	return diffService
}

func (*DiffService) Diff(old, new *html.Node) string {
	diffs, err := compareTrees(old, new, "")
	if err != nil {
		return ""
	}

	return diffsToString(diffs)
}

func compareTrees(n1, n2 *html.Node, path string) ([]Diff, error) {
	diffs := []Diff{}

	// Compare current nodes
	if n1 == nil && n2 == nil {
		return diffs, nil
	}

	if n1 == nil {
		diffs = append(diffs, Diff{
			Type:    Added,
			Path:    path,
			New:     nodeToString(n2),
			Message: fmt.Sprintf("Node added at %s", path),
		})
		return diffs, nil
	}

	if n2 == nil {
		diffs = append(diffs, Diff{
			Type:    Removed,
			Path:    path,
			Old:     nodeToString(n1),
			Message: fmt.Sprintf("Node removed at %s", path),
		})
		return diffs, nil
	}

	// Compare node type
	if n1.Type != n2.Type {
		diffs = append(diffs, Diff{
			Type:    Modified,
			Path:    path,
			Old:     fmt.Sprintf("Type: %v", n1.Type),
			New:     fmt.Sprintf("Type: %v", n2.Type),
			Message: fmt.Sprintf("Node type changed at %s", path),
		})
		return diffs, nil
	}

	// Compare element nodes
	if n1.Type == html.ElementNode {
		newPath := path + "/" + n1.Data

		// Compare tag names
		if n1.Data != n2.Data {
			diffs = append(diffs, Diff{
				Type:    Modified,
				Path:    newPath,
				Old:     n1.Data,
				New:     n2.Data,
				Message: fmt.Sprintf("Tag changed from <%s> to <%s>", n1.Data, n2.Data),
			})
		}

		// Compare attributes
		attrDiffs, err := compareAttributes(n1, n2, newPath)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, attrDiffs...)

		// Compare children
		childDiffs, err := compareChildren(n1, n2, newPath)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, childDiffs...)
	} else if n1.Type == html.TextNode {
		// Compare text content (trim whitespace for comparison)
		text1 := strings.TrimSpace(n1.Data)
		text2 := strings.TrimSpace(n2.Data)

		if text1 != text2 && (text1 != "" || text2 != "") {
			diffs = append(diffs, Diff{
				Type:    Modified,
				Path:    path,
				Old:     text1,
				New:     text2,
				Message: fmt.Sprintf("Text changed at %s", path),
			})
		}

		// Still compare siblings
		siblingDiffs, err := compareTrees(n1.NextSibling, n2.NextSibling, path)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, siblingDiffs...)
	} else {
		// For other node types, compare siblings
		siblingDiffs, err := compareTrees(n1.NextSibling, n2.NextSibling, path)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, siblingDiffs...)
	}

	return diffs, nil
}

// compareAttributes compares attributes of two element nodes
func compareAttributes(n1, n2 *html.Node, path string) ([]Diff, error) {
	diffs := []Diff{}
	attrs1 := make(map[string]string)
	attrs2 := make(map[string]string)

	for _, attr := range n1.Attr {
		attrs1[attr.Key] = attr.Val
	}

	for _, attr := range n2.Attr {
		attrs2[attr.Key] = attr.Val
	}

	// Check for removed or modified attributes
	for key, val1 := range attrs1 {
		if val2, exists := attrs2[key]; !exists {
			diffs = append(diffs, Diff{
				Type:    Removed,
				Path:    path + "[@" + key + "]",
				Old:     val1,
				Message: fmt.Sprintf("Attribute '%s' removed", key),
			})
		} else if val1 != val2 {
			diffs = append(diffs, Diff{
				Type:    Modified,
				Path:    path + "[@" + key + "]",
				Old:     val1,
				New:     val2,
				Message: fmt.Sprintf("Attribute '%s' changed from %s to %s", key, val1, val2),
			})
		}
	}

	// Check for added attributes
	for key, val2 := range attrs2 {
		if _, exists := attrs1[key]; !exists {
			diffs = append(diffs, Diff{
				Type:    Added,
				Path:    path + "[@" + key + "]",
				New:     val2,
				Message: fmt.Sprintf("Attribute '%s' added", key),
			})
		}
	}

	return diffs, nil
}

// compareChildren compares child nodes
func compareChildren(n1, n2 *html.Node, path string) ([]Diff, error) {
	diffs := []Diff{}
	c1 := n1.FirstChild
	c2 := n2.FirstChild

	index := 0
	for c1 != nil || c2 != nil {
		childPath := fmt.Sprintf("%s[%d]", path, index)
		childDiffs, err := compareTrees(c1, c2, childPath)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, childDiffs...)

		if c1 != nil {
			c1 = c1.NextSibling
		}
		if c2 != nil {
			c2 = c2.NextSibling
		}
		index++
	}

	return diffs, nil
}

func nodeToString(n *html.Node) string {
	if n == nil {
		return ""
	}
	switch n.Type {
	case html.ElementNode:
		return fmt.Sprintf("<%s>", n.Data)
	case html.TextNode:
		return strings.TrimSpace(n.Data)
	default:
		return fmt.Sprintf("Node(%v)", n.Type)
	}
}
func diffsToString(diffs []Diff) string {
	var result strings.Builder
	for i, diff := range diffs {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, diff.Message))
	}
	return result.String()
}
