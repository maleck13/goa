package design

import "github.com/goadesign/goa/eval"
import "github.com/goadesign/goa/design"

type (
	// FileServerExpression defines an endpoint that servers static assets.
	FileServerExpression struct {
		// Parent resource
		Parent *ResourceExpr
		// Description for docs
		Description string
		// Docs points to the service external documentation
		Docs *design.DocsExpr
		// FilePath is the file path to the static asset(s)
		FilePath string
		// RequestPath is the HTTP path that servers the assets.
		RequestPath string
		// Metadata is a list of key/value pairs
		Metadata eval.MetadataExpr
	}
)
