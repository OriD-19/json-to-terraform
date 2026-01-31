package parser

// Options configures the parser behavior.
type Options struct {
	// EmitTfvars generates terraform.tfvars from diagram metadata when true.
	EmitTfvars bool
	// MaxParallel is the max number of nodes to process in parallel per tier (0 = default).
	MaxParallel int
}

// DefaultOptions returns default parser options.
func DefaultOptions() Options {
	return Options{
		EmitTfvars:  true,
		MaxParallel: 0, // use runtime.NumCPU in parser
	}
}
