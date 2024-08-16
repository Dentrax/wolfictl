package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/dot"
	"github.com/wolfi-dev/wolfictl/pkg/question"
)

// Dot generates a DOT graph from the given root question and returns it as a
// string. The graph will be directed and will represent the flow of questions
// and choices in the interview.
func Dot[T any](ctx context.Context, root question.Question[T], initialState T) (string, error) {
	g := dot.NewGraph("interview")
	g.SetType(dot.DIGRAPH) //nolint:errcheck

	// Create a "Done" node
	doneNode := dot.NewNode("Done")
	g.AddNode(doneNode) //nolint:errcheck

	startNode := dot.NewNode(fmt.Sprintf(`"%+v"`, initialState))

	err := traverse(ctx, g, root, initialState, startNode, "", doneNode)
	if err != nil {
		return "", fmt.Errorf("traversing graph: %w", err)
	}

	return g.String(), nil
}

func traverse[T any](
	ctx context.Context,
	g *dot.Graph,
	q question.Question[T],
	state T,
	parentNode *dot.Node,
	edgeLabel string,
	doneNode *dot.Node,
) error {
	// Create a unique ID for this question based on its text
	id := fmt.Sprintf(`"%s"`, q.Text)

	// Add this question as a node in the graph
	node := dot.NewNode(id)
	g.AddNode(node) //nolint:errcheck

	// If this question has a parent, add an edge from the parent to this question
	if parentNode != nil {
		edge := dot.NewEdge(parentNode, node)
		_ = edge.Set("label", edgeLabel) //nolint:errcheck
		g.AddEdge(edge)                  //nolint:errcheck
	}

	switch a := q.Answer.(type) {
	case question.AcceptText[T]:
		// Simulate a text answer to advance the propagation through the graph.
		answer := textInput

		updatedState, nextQuestion, err := a(state, answer)
		if err != nil {
			if errors.Is(err, question.ErrTerminate) {
				connectToTerminatedNode(g, node, answer)
				return nil
			}

			return err
		}
		if nextQuestion != nil {
			err := traverse(ctx, g, *nextQuestion, updatedState, node, answer, doneNode)
			if err != nil {
				return err
			}
			return nil
		}

		// If the choice leads to a nil question, create an edge to the "Done" node
		edge := dot.NewEdge(node, doneNode)
		_ = edge.Set("label", answer) //nolint:errcheck
		g.AddEdge(edge)               //nolint:errcheck

	case question.MultipleChoice[T]:
		// Iterate over the choices for this question
		for _, choice := range a {
			if choice.Choose == nil {
				continue
			}

			// If the choice leads to another question, recursively traverse that question
			updatedState, nextQuestion, err := choice.Choose(state)
			if err != nil {
				if errors.Is(err, question.ErrTerminate) {
					connectToTerminatedNode(g, node, choice.Text)
					return nil
				}

				return err
			}
			if nextQuestion != nil {
				err := traverse(ctx, g, *nextQuestion, updatedState, node, choice.Text, doneNode)
				if err != nil {
					return err
				}
				continue
			}

			// If the choice leads to a nil question, create an edge to the "Done" node
			edge := dot.NewEdge(node, doneNode)
			_ = edge.Set("label", choice.Text) //nolint:errcheck
			g.AddEdge(edge)                    //nolint:errcheck
		}

	case question.MessageOnly[T]:
		edgeLabel := messageAccepted
		updatedState, nextQuestion, err := a(state)
		if err != nil {
			if errors.Is(err, question.ErrTerminate) {
				connectToTerminatedNode(g, node, edgeLabel)
				return nil
			}

			return err
		}
		if nextQuestion != nil {
			err := traverse(ctx, g, *nextQuestion, updatedState, node, edgeLabel, doneNode)
			if err != nil {
				return err
			}
			return nil
		}

		// If the choice leads to a nil question, create an edge to the "Done" node
		edge := dot.NewEdge(node, doneNode)
		_ = edge.Set("label", edgeLabel) //nolint:errcheck
		g.AddEdge(edge)                  //nolint:errcheck
	}

	return nil
}

func connectToTerminatedNode(g *dot.Graph, node *dot.Node, edgeLabel string) {
	term := terminatedNode()
	g.AddNode(term) //nolint:errcheck
	edge := dot.NewEdge(node, term)
	_ = edge.Set("label", edgeLabel) //nolint:errcheck
	g.AddEdge(edge)                  //nolint:errcheck
}
func terminatedNode() *dot.Node {
	return dot.NewNode(`"<EXIT WITH NO RESULT>"`)
}

const (
	textInput       = "<TEXT INPUT>"
	messageAccepted = "<MESSAGE ACCEPTED>"
)
