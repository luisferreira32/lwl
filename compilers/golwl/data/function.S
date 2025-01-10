# Filename: function.S

.section .text
    .global _start      # Make _start visible to the linker

f:
    mov  %rdi, %rax     # Move first parameter (from RDI) into the return register, RAX.
    add  %rsi, %rax     # Add second parameter (from RSI) to RAX. Now RAX = a + b.
    ret                 # Return from function.

# Program entry point
_start:
    # Call f(1, 2)
    mov  $1, %rdi       # Set the first argument to 1.
    mov  $2, %rsi       # Set the second argument to 2.
    call f              # Call the function 'f'.

    # Exit the program with the return code from 'f'.
    mov  %rax, %rdi     # The exit system call expects the exit code in RDI.
    mov  $60, %rax      # 60 is the system call number for 'exit'.
    syscall             # Kernel, do the system call!