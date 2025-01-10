# ################################################################################## #
#  Variables                                                                         #
# ################################################################################## #
GO            := go
VERSION       := $(shell cat VERSION)
BIN           := ./bin/golwl
TEST_BIN      := ./bin/golwl.test
GCC           := gcc
GCC_C_FLAGS   := -O0
LD            := ld
OBJDUMP       := objdump
OBJDUMP_FLAGS := -d -M intel

C_SOURCES     := $(wildcard data/*.c)
S_SOURCES   := $(wildcard data/*.S)
C_S_TARGETS   := $(patsubst data/%.c, out/%.c.s, $(C_SOURCES))
C_O_TARGETS   := $(patsubst data/%.c, out/%.c.o.objdump, $(C_SOURCES))
C_B_TARGETS   := $(patsubst data/%.c, out/%.c.bin.objdump, $(C_SOURCES))
S_O_TARGETS := $(patsubst data/%.S, out/%.S.o.objdump, $(S_SOURCES))
S_B_TARGETS := $(patsubst data/%.S, out/%.S.bin.objdump, $(S_SOURCES))

# ################################################################################## #
#  Phonies                                                                           #
# ################################################################################## #
.PHONY: all
all: compile gen

.PHONY: compile
compile: bin/golwl

.PHONY: test
test: bin/golwl bin/golwl.test

.PHONY: gen
gen: $(C_S_TARGETS) $(C_O_TARGETS) $(C_B_TARGETS) $(S_O_TARGETS) $(S_B_TARGETS)

.PHONY: clean
clean:
	rm bin/*
	rm out/*

# ################################################################################## #
#  Go LWL compiler targets                                                           #
# ################################################################################## #
bin/golwl: *.go
	$(GO) build -ldflags "-X main.version=${VERSION}" -o ${BIN} .

bin/golwl.test: *_test.go
	$(GO) test -c -o ${TEST_BIN} .

# ################################################################################## #
#  Test data generation targets                                                      #
# ################################################################################## #
out/%.c.o: data/%.c
	$(GCC) $(GCC_C_FLAGS) -c $< -o $@

out/%.c.s: data/%.c
	$(GCC) -S $< -o $@

out/%.c.o.objdump: out/%.c.o
	$(OBJDUMP) $(OBJDUMP_FLAGS) $< > $@

out/%.c.bin: out/%.c.o
	$(GCC) $< -o $@

out/%.c.bin.objdump: out/%.c.bin
	$(OBJDUMP) $(OBJDUMP_FLAGS) $< > $@

out/%.S.o: data/%.S
	$(GCC) -c $< -o $@

out/%.S.o.objdump: out/%.S.o
	$(OBJDUMP) $(OBJDUMP_FLAGS) $< > $@

out/%.S.bin: out/%.S.o
	ld -o $@ $<

out/%.S.bin.objdump: out/%.S.bin
	$(OBJDUMP) $(OBJDUMP_FLAGS) $< > $@
