package anonymizer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/planforge/internal/parser"
)

var (
	reStringLiteral = regexp.MustCompile(`'[^']*'`)
	reNumberLiteral = regexp.MustCompile(`\b\d+(\.\d+)?\b`)
)

// Sanitizer provides anonymization mapping for execution plans.
type Sanitizer struct {
	tableMap  map[string]string
	columnMap map[string]string
	indexMap  map[string]string
	tableSeq  int
	colSeq    int
	indexSeq  int
}

// NewSanitizer creates a new sanitizer instance.
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		tableMap:  make(map[string]string),
		columnMap: make(map[string]string),
		indexMap:  make(map[string]string),
	}
}

// AnonymizePlan clones and obfuscates relations, columns, and filter literals.
func (s *Sanitizer) AnonymizePlan(orig *parser.Plan) *parser.Plan {
	if orig == nil {
		return nil
	}

	clone := *orig
	if orig.Root != nil {
		clone.Root = s.anonymizeNode(orig.Root)
	}

	clone.RawInput = ""
	return &clone
}

func (s *Sanitizer) anonymizeNode(n *parser.PlanNode) *parser.PlanNode {
	if n == nil {
		return nil
	}

	newNode := *n

	// Obfuscate relation
	if n.RelationName != "" {
		if _, exists := s.tableMap[n.RelationName]; !exists {
			s.tableSeq++
			s.tableMap[n.RelationName] = fmt.Sprintf("tbl_%d", s.tableSeq)
		}
		newNode.RelationName = s.tableMap[n.RelationName]
	}

	if n.Alias != "" {
		newNode.Alias = fmt.Sprintf("alias_%s", newNode.RelationName)
	}

	// Obfuscate index name
	if n.IndexName != "" {
		if _, exists := s.indexMap[n.IndexName]; !exists {
			s.indexSeq++
			s.indexMap[n.IndexName] = fmt.Sprintf("idx_rel_%d", s.indexSeq)
		}
		newNode.IndexName = s.indexMap[n.IndexName]
	}

	// Obfuscate Filter and IndexCond expressions
	newNode.Filter = s.maskExpression(n.Filter)
	newNode.IndexCond = s.maskExpression(n.IndexCond)
	newNode.HashCond = s.maskExpression(n.HashCond)
	newNode.MergeCond = s.maskExpression(n.MergeCond)

	// Obfuscate Output columns
	if len(n.Output) > 0 {
		var newOut []string
		for _, col := range n.Output {
			newOut = append(newOut, s.maskIdentifier(col))
		}
		newNode.Output = newOut
	}

	// Obfuscate Sort Keys
	if len(n.SortKey) > 0 {
		var newSort []string
		for _, sk := range n.SortKey {
			newSort = append(newSort, s.maskExpression(sk))
		}
		newNode.SortKey = newSort
	}

	// Recursively anonymize children
	if len(n.Children) > 0 {
		newNode.Children = make([]*parser.PlanNode, len(n.Children))
		for i, child := range n.Children {
			newNode.Children[i] = s.anonymizeNode(child)
		}
	}

	return &newNode
}

func (s *Sanitizer) maskExpression(expr string) string {
	if expr == "" {
		return ""
	}

	// Mask strings
	res := reStringLiteral.ReplaceAllString(expr, "'***'")

	// Replace known table names
	for origTable, maskTable := range s.tableMap {
		res = strings.ReplaceAll(res, origTable, maskTable)
	}

	return res
}

func (s *Sanitizer) maskIdentifier(name string) string {
	if name == "" {
		return ""
	}
	if _, exists := s.columnMap[name]; !exists {
		s.colSeq++
		s.columnMap[name] = fmt.Sprintf("col_%d", s.colSeq)
	}
	return s.columnMap[name]
}
