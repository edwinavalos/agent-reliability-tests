#!/usr/bin/env python3
"""
Multi-Agent Communication Reliability Test Coordinator
Orchestrates 100-loop communication test between python-pro and fullstack-developer agents
"""

import os
import time
import subprocess
import datetime
from pathlib import Path

class AgentCoordinator:
    def __init__(self, test_dir="/home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests"):
        self.test_dir = Path(test_dir)
        self.message_to_fullstack = self.test_dir / "message_to_fullstack.txt"
        self.message_to_python = self.test_dir / "message_to_python.txt"
        self.chat_log = self.test_dir / "chat.log"
        self.successful_cycles = 0
        self.failed_cycles = 0
        self.errors = []
        
    def log_event(self, message):
        """Append event to chat.log with timestamp"""
        timestamp = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]
        with open(self.chat_log, 'a') as f:
            f.write(f"[{timestamp}] {message}\n")
    
    def clean_message_files(self):
        """Clean message files for next iteration"""
        try:
            self.message_to_fullstack.write_text("")
            self.message_to_python.write_text("")
            return True
        except Exception as e:
            self.errors.append(f"Failed to clean message files: {e}")
            return False
    
    def wait_for_file_content(self, file_path, expected_content=None, timeout=30):
        """Wait for file to have content or specific content"""
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                if file_path.exists():
                    content = file_path.read_text().strip()
                    if content:
                        if expected_content is None or expected_content in content:
                            return content
                time.sleep(0.1)
            except Exception as e:
                self.errors.append(f"Error reading {file_path}: {e}")
        return None
    
    def launch_python_pro(self, loop_num):
        """Launch python-pro agent with communication instructions"""
        instructions = f"""
You are participating in loop {loop_num} of a 100-loop communication reliability test.

EXACT STEPS TO FOLLOW:
1. Write "hello" to /home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_fullstack.txt
2. Append "python-pro: hello" to /home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/chat.log  
3. Wait for and read response from /home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_python.txt
4. Once you receive the response, your part is complete

CRITICAL: Use actual file operations, not simulation. The fullstack-developer agent is waiting for your message.
"""
        
        try:
            # Launch python-pro via Claude Code CLI
            cmd = ["claude-code", "--agent", "python-pro", "--message", instructions]
            process = subprocess.Popen(cmd, cwd=self.test_dir, capture_output=True, text=True)
            return process
        except Exception as e:
            self.errors.append(f"Failed to launch python-pro in loop {loop_num}: {e}")
            return None
    
    def launch_fullstack_developer(self, loop_num):
        """Launch fullstack-developer agent with communication instructions"""
        instructions = f"""
You are participating in loop {loop_num} of a 100-loop communication reliability test.

EXACT STEPS TO FOLLOW:
1. Read the message from /home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_fullstack.txt
2. Write "world" to /home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/message_to_python.txt
3. Append "fullstack-developer: world" to /home/edwin/GolandProjects/PersonalClaude/agent-reliability-tests/chat.log
4. Signal completion

CRITICAL: Use actual file operations, not simulation. The python-pro agent has sent you a message.
"""
        
        try:
            # Launch fullstack-developer via Claude Code CLI  
            cmd = ["claude-code", "--agent", "fullstack-developer", "--message", instructions]
            process = subprocess.Popen(cmd, cwd=self.test_dir, capture_output=True, text=True)
            return process
        except Exception as e:
            self.errors.append(f"Failed to launch fullstack-developer in loop {loop_num}: {e}")
            return None
    
    def run_communication_loop(self, loop_num):
        """Execute one complete communication loop"""
        self.log_event(f"=== Starting Communication Loop {loop_num} ===")
        
        # Clean message files
        if not self.clean_message_files():
            self.failed_cycles += 1
            return False
        
        # Step 1: Launch python-pro
        self.log_event(f"Loop {loop_num}: Launching python-pro agent")
        python_process = self.launch_python_pro(loop_num)
        if not python_process:
            self.failed_cycles += 1
            return False
        
        # Step 2: Wait for python-pro to write message
        self.log_event(f"Loop {loop_num}: Waiting for python-pro message")
        python_message = self.wait_for_file_content(self.message_to_fullstack, "hello", 60)
        if not python_message:
            self.log_event(f"Loop {loop_num}: TIMEOUT - python-pro message not received")
            python_process.terminate()
            self.failed_cycles += 1
            return False
        
        self.log_event(f"Loop {loop_num}: Received python-pro message: {python_message}")
        
        # Step 3: Launch fullstack-developer  
        self.log_event(f"Loop {loop_num}: Launching fullstack-developer agent")
        fullstack_process = self.launch_fullstack_developer(loop_num)
        if not fullstack_process:
            python_process.terminate()
            self.failed_cycles += 1
            return False
        
        # Step 4: Wait for fullstack-developer to respond
        self.log_event(f"Loop {loop_num}: Waiting for fullstack-developer response")
        fullstack_response = self.wait_for_file_content(self.message_to_python, "world", 60)
        if not fullstack_response:
            self.log_event(f"Loop {loop_num}: TIMEOUT - fullstack-developer response not received")
            python_process.terminate()
            fullstack_process.terminate()
            self.failed_cycles += 1
            return False
        
        self.log_event(f"Loop {loop_num}: Received fullstack-developer response: {fullstack_response}")
        
        # Step 5: Wait for processes to complete
        try:
            python_process.wait(timeout=30)
            fullstack_process.wait(timeout=30)
        except subprocess.TimeoutExpired:
            self.log_event(f"Loop {loop_num}: Agent processes timed out")
            python_process.terminate()
            fullstack_process.terminate()
            self.failed_cycles += 1
            return False
        
        # Step 6: Verify communication completed successfully
        self.log_event(f"Loop {loop_num}: Communication cycle completed successfully")
        self.successful_cycles += 1
        return True
    
    def run_full_test(self):
        """Execute the complete 100-loop test"""
        self.log_event("Starting 100-loop agent communication reliability test")
        
        start_time = time.time()
        
        for loop_num in range(1, 101):
            try:
                success = self.run_communication_loop(loop_num)
                if not success:
                    self.log_event(f"Loop {loop_num}: FAILED")
                
                # Brief pause between loops
                time.sleep(1)
                
            except KeyboardInterrupt:
                self.log_event("Test interrupted by user")
                break
            except Exception as e:
                self.log_event(f"Loop {loop_num}: Unexpected error: {e}")
                self.errors.append(f"Loop {loop_num} unexpected error: {e}")
                self.failed_cycles += 1
        
        end_time = time.time()
        duration = end_time - start_time
        
        # Generate final report
        self.generate_final_report(duration)
    
    def generate_final_report(self, duration):
        """Generate comprehensive test results report"""
        # Count chat.log lines
        try:
            with open(self.chat_log, 'r') as f:
                chat_log_lines = len(f.readlines())
        except:
            chat_log_lines = "Unknown"
        
        # Calculate reliability metrics
        total_attempts = self.successful_cycles + self.failed_cycles
        reliability_rate = (self.successful_cycles / total_attempts * 100) if total_attempts > 0 else 0
        
        report = f"""

=== AGENT COMMUNICATION RELIABILITY TEST RESULTS ===
Test Duration: {duration:.2f} seconds
Total Communication Cycles Attempted: {total_attempts}
Successful Communication Cycles: {self.successful_cycles}
Failed Communication Cycles: {self.failed_cycles}
Communication Reliability Rate: {reliability_rate:.2f}%
Chat Log Final Line Count: {chat_log_lines}

=== ERROR SUMMARY ===
Total Errors Encountered: {len(self.errors)}
"""
        
        if self.errors:
            report += "\nDetailed Errors:\n"
            for i, error in enumerate(self.errors, 1):
                report += f"{i}. {error}\n"
        else:
            report += "No errors encountered during testing.\n"
        
        report += f"""
=== ASSESSMENT ===
Inter-agent communication reliability: {'EXCELLENT' if reliability_rate >= 95 else 'GOOD' if reliability_rate >= 85 else 'POOR'}
File-based message passing protocol: {'RELIABLE' if self.failed_cycles < 5 else 'UNRELIABLE'}
Agent coordination effectiveness: {'HIGH' if self.successful_cycles >= 90 else 'MODERATE' if self.successful_cycles >= 70 else 'LOW'}

Test completed at: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
"""
        
        # Log the report
        self.log_event(report)
        
        # Also print to console
        print(report)

if __name__ == "__main__":
    coordinator = AgentCoordinator()
    coordinator.run_full_test()