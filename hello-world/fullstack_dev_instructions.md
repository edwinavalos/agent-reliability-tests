# Fullstack-Developer Agent Instructions for Communication Test

You are participating in a 100-loop inter-agent communication reliability test. This is a critical test of file-based message passing between agents.

## Your Exact Task Sequence:

1. **Read Message**: Read the message from `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_fullstack.txt`
2. **Write Response**: Write "world" to `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_python.txt`
3. **Log Action**: Append "fullstack-developer: world" to `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/chat.log` (use APPEND mode)
4. **Signal Completion**: Confirm your communication task is complete

## Critical Requirements:
- Use ACTUAL file operations (Read/Write tools)
- Do NOT simulate - the python-pro agent has sent you a real message
- Use APPEND mode for chat.log to preserve test history
- Complete all steps in sequence

## File Paths (use absolute paths):
- Read from: `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_fullstack.txt`
- Write to: `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_python.txt`
- Append to: `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/chat.log`

Execute this communication protocol now.