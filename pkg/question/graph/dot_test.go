package graph

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wolfi-dev/wolfictl/pkg/question"
)

func TestDot(t *testing.T) {
	var (
		qIcecreamFlavor = question.Question[string]{
			Text: "What flavor of ice cream do you like?",
			Answer: question.MultipleChoice[string]{
				{
					Text: "Vanilla",
					Choose: func(state string) (string, *question.Question[string], error) {
						return "vanilla " + state, nil, nil
					},
				},
				{
					Text: "Chocolate",
					Choose: func(state string) (string, *question.Question[string], error) {
						return "chocolate " + state, nil, nil
					},
				},
			},
		}

		qCookieKind = question.Question[string]{
			Text: "What kind of cookie do you like?",
			Answer: question.MultipleChoice[string]{
				{
					Text: "Chocolate chip",
					Choose: func(state string) (string, *question.Question[string], error) {
						return "chocolate chip " + state, nil, nil
					},
				},
			},
		}

		qFavoriteDessert = question.Question[string]{
			Text: "What is your favorite dessert?",
			Answer: question.MultipleChoice[string]{
				{
					Text: "Ice cream",
					Choose: func(_ string) (string, *question.Question[string], error) {
						return "ice cream", &qIcecreamFlavor, nil
					},
				},
				{
					Text: "Cookie",
					Choose: func(_ string) (string, *question.Question[string], error) {
						return "cookie", &qCookieKind, nil
					},
				},
			},
		}
	)

	var (
		expected = `digraph interview {
Done;
"What is your favorite dessert?";
"START" -> "What is your favorite dessert?" [ label="" ]
"What flavor of ice cream do you like?";
"What is your favorite dessert?" -> "What flavor of ice cream do you like?" [ label="Ice cream" ]
"What flavor of ice cream do you like?" -> Done [ label=Vanilla ]
"What flavor of ice cream do you like?" -> Done [ label=Chocolate ]
"What kind of cookie do you like?";
"What is your favorite dessert?" -> "What kind of cookie do you like?" [ label=Cookie ]
"What kind of cookie do you like?" -> Done [ label="Chocolate chip" ]
}
`
	)

	dot, err := Dot(context.Background(), qFavoriteDessert, "START")
	if err != nil {
		t.Fatalf("Dot() error = %v", err)
	}

	if diff := cmp.Diff(expected, dot); diff != "" {
		t.Errorf("Dot() unexpected output (-want +got):\n%s", diff)
	}
}
