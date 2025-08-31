# Example Prompt Templates

This directory contains example Go template files for use with the agent-reliability-tests tool.

## Template Format

Templates use Go's `text/template` syntax. The available variables are:
- `{{.SubAgentName}}` - The name of the agent being tested

## Template Files

### hello_world.tmpl
The default template - asks the agent to say hello and return the response.

### coordination_plan.tmpl  
Tests multi-agent coordination by asking the agent to create an implementation plan using subagents.

### code_review.tmpl
Tests code review capabilities by asking the agent to review a simple Go program.

### feature_implementation.tmpl
Tests feature implementation by asking the agent to build a complete REST API with authentication.

## Usage

```bash
# Use a custom template
./agent-reliability-tests general-purpose --prompt example_prompt_templates/coordination_plan.tmpl

# Use with multiple workers
./agent-reliability-tests multi-agent-coordinator --prompt example_prompt_templates/coordination_plan.tmpl --queue 3 --loops 5
```

## File Extensions

Template files must use either `.tmpl` or `.template` extensions to be accepted by the validation system.