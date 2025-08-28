# Python-Pro Agent Instructions for Communication Test

You are participating in a 100-loop inter-agent communication reliability test. This is a critical test of file-based message passing between agents.

## Your Exact Task Sequence:

1. **Write Message**: Write "hello" to `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_fullstack.txt`
2. **Log Action**: Append "python-pro: hello" to `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/chat.log` (use APPEND mode)
3. **Wait for Response**: Monitor `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_python.txt` until it contains content
4. **Confirm Receipt**: Read the response and confirm the communication cycle is complete

## Critical Requirements:
- Use ACTUAL file operations (Read/Write tools)
- Do NOT simulate - the fullstack-developer agent is actually waiting
- Use APPEND mode for chat.log to preserve test history
- Wait for actual response before completing

## File Paths (use absolute paths):
- Write to: `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_fullstack.txt`
- Read from: `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_python.txt`
- Append to: `/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/chat.log`

Execute this communication protocol now.