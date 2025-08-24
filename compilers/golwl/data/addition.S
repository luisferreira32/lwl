# Filename: addition.S

.section .text

.global _start      # Make _start visible to the linker

# Program entry point
_start:
    mov  $1, %rdi       # Set the first argument to 1.
    mov  $3, %rax       # Set the second argument to 3.
    add  %rdi, %rax     # Add second parameter (from RSI)
    mov  $1, %rdi       # Set the first argument to 1.
    add  %rdi, %rax     # Add second parameter (from RSI)
    mov  %rax, %rdi     # The exit system call expects the exit code in RDI.
    mov  $60, %rax      # 60 is the system call number for 'exit'.
    syscall             # Kernel, do the system call!