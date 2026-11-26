// Package link answers whether a connection exists between two node records. It reads records
// and adds nothing to them, so it is a layer above the node model rather than part of it. Nothing
// is kept between builds: an answer is read from the records at the moment it is asked.
package link

import (
	"sort"

	"github.com/Galdoba/go-graph/pkg/graph"
)

// Pair is a connection as it is answered: the two records it is about, in outgoing order.
type Pair struct {
	Source graph.ID `json:"source"`
	Target graph.ID `json:"target"`
}

// UndecidedError reports that a rule could not answer for a pair, and it carries the name of
// the rule. Nothing else fails, because a rule reads only records.
type UndecidedError struct {
	Rule string `json:"rule"`
}

// Error names the rule that could not answer, which is all the third outcome carries.
func (e UndecidedError) Error() string { return e.Rule }

// Rule states one relation type and answers for one pair using only what the two members of the
// pair carry. The rule name it states is passed to its answer, which is what the third outcome
// reports.
type Rule struct {
	Name   string
	Answer func(name string, source, target *graph.Node) (bool, error)
}

// parentMethod is the relation from a structure to the method that records it as its parent.
var parentMethod = Rule{
	Name:   "structure to the method whose parent it is",
	Answer: answerParentMethod,
}

// referentNode is the relation from a structure to a node one of its attributes refers to.
var referentNode = Rule{
	Name:   "structure to a node one of its attributes refers to",
	Answer: answerReferentNode,
}

// satisfiedInterface is the relation from an interface to the structure that records every method
// the interface states.
var satisfiedInterface = Rule{
	Name:   "interface to a struct that satisfies it",
	Answer: answerSatisfiedInterface,
}

// carriedReference is the relation from a node to the node a reference it carries names.
var carriedReference = Rule{
	Name:   "node to a node a reference it carries names",
	Answer: answerCarriedReference,
}

// rules is the closed set of rules, in the order they are consulted. The set grows only by adding
// a rule.
var rules = []Rule{parentMethod, referentNode, satisfiedInterface, carriedReference}

// Exists answers one pair against the rules in the order they are consulted, stopping at the first
// rule that answers yes. When no rule answers yes, the outcome is that the connection does not
// exist only when every rule that could answer answered no.
func Exists(source, target *graph.Node) (bool, error) {
	return consult(rules, source, target)
}

// Pairs returns the connections the rules answer for the records in the collection, ordered by the
// source identifier and then by the target identifier. The answer is read from the records at the
// moment it is asked, and nothing is kept between calls.
func Pairs(nodes []*graph.Node) []Pair {
	out := []Pair{}
	for _, source := range nodes {
		// A pair has two different records as its members, so a record is never a pair with itself.
		for _, target := range nodes {
			if source.ID == target.ID {
				continue
			}
			yes, err := consult(rules, source, target)
			if yes && err == nil {
				out = append(out, Pair{Source: source.ID, Target: target.ID})
			}
		}
	}
	// A set of pairs is not a set of attributes, so ordering by the pair is what makes the set
	// reproducible.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source.String() != out[j].Source.String() {
			return out[i].Source.String() < out[j].Source.String()
		}
		return out[i].Target.String() < out[j].Target.String()
	})
	return out
}

// Coverage reports, in the order they are consulted, the rules that no pair of the collection
// answers yes for. A rule answered yes by no pair is reported as not covered.
func Coverage(rs []Rule, nodes []*graph.Node) []string {
	out := []string{}
	for _, r := range rs {
		covered := false
		// A rule is exercised by the test ground when one of its pairs is answered yes on real code.
		for _, source := range nodes {
			for _, target := range nodes {
				if source.ID == target.ID {
					continue
				}
				yes, err := r.Answer(r.Name, source, target)
				if yes && err == nil {
					covered = true
				}
			}
		}
		if !covered {
			out = append(out, r.Name)
		}
	}
	return out
}

func answerParentMethod(name string, source, target *graph.Node) (bool, error) {
	if source.SemanticType == graph.SemanticUnresolved || target.SemanticType == graph.SemanticUnresolved {
		// A rule that reads a semantic value cannot answer for an unresolved one, so the outcome is
		// the third one rather than no.
		return false, &UndecidedError{Rule: name}
	}
	if source.SemanticType != graph.SemanticStructure || target.SemanticType != graph.SemanticMethod {
		// A rule that requires a structure as the source answers no for a source it does not cover.
		return false, nil
	}
	return source.ID == target.Parent, nil
}

func answerReferentNode(name string, source, target *graph.Node) (bool, error) {
	if source.SemanticType == graph.SemanticUnresolved || target.SemanticType == graph.SemanticUnresolved {
		return false, &UndecidedError{Rule: name}
	}
	if source.SemanticType != graph.SemanticStructure {
		return false, nil
	}
	for _, a := range source.Attributes {
		// A method attribute records the method set of the structure, not a node the structure
		// refers to, so it names no relation of this kind.
		if a.Kind == graph.AttrMethod {
			continue
		}
		// An attribute whose reference is the record that owns it names no relation between two
		// records. A reference with no name is not a referent node.
		if a.Ref.Name != "" && a.Ref != source.ID && a.Ref == target.ID {
			return true, nil
		}
	}
	return false, nil
}

func answerSatisfiedInterface(name string, source, target *graph.Node) (bool, error) {
	if source.SemanticType == graph.SemanticUnresolved || target.SemanticType == graph.SemanticUnresolved {
		return false, &UndecidedError{Rule: name}
	}
	if source.SemanticType != graph.SemanticInterface || target.SemanticType != graph.SemanticStructure {
		return false, nil
	}
	specs := 0
	for _, a := range source.Attributes {
		if a.Kind != graph.AttrMethodSpec {
			continue
		}
		specs++
		if !recordsMethod(target, a) {
			return false, nil
		}
	}
	// An interface with no method spec states nothing to check, so this rule answers no.
	return specs > 0, nil
}

// recordsMethod reports whether the structure carries the method a method spec states. The method
// is recorded on the node that owns it, so being recorded on the structure is the receiver, and the
// signature is compared as recorded, without the receiver.
func recordsMethod(target *graph.Node, spec graph.Attribute) bool {
	for _, a := range target.Attributes {
		if a.Kind == graph.AttrMethod && a.Name == spec.Name && a.Value == spec.Value {
			return true
		}
	}
	return false
}

// answerCarriedReference reads the references the source carries, which is all a pair can hold for
// this relation. Because it reads no semantic value, it cannot be unable to answer for an
// unresolved record. A record read from a name states no uses, because no declaration body of that
// package was read, so this rule cannot read an absence as a claim that the entity refers to
// nothing.
func answerCarriedReference(name string, source, target *graph.Node) (bool, error) {
	if source.Provenance == graph.ProvenanceFromName {
		return false, &UndecidedError{Rule: name}
	}
	for _, a := range source.Attributes {
		// Only a use records an entity a declaration refers to, so an alias or a method attribute
		// names no connection of this kind. A reference with no name points at no record.
		if a.Kind != graph.AttrUse || a.Ref.Name == "" {
			continue
		}
		if a.Ref == target.ID {
			return true, nil
		}
	}
	return false, nil
}

// consult answers one pair against the rules in order. The consultation stops at the first yes.
// The rule reported for the undecided outcome is the last one in that order that could not answer.
func consult(rs []Rule, source, target *graph.Node) (bool, error) {
	last := ""
	for _, r := range rs {
		yes, err := r.Answer(r.Name, source, target)
		if yes {
			return true, nil
		}
		// The reported rule is the last one that could not answer, overwriting the earlier ones.
		if err != nil {
			last = r.Name
		}
	}
	if last != "" {
		return false, &UndecidedError{Rule: last}
	}
	return false, nil
}
