.section .text
_start:
	addi sp, sp, -16
	addi t0, zero, 72
	sb t0, 0(sp)
	addi t0, zero, 105
	sb t0, 1(sp)
	addi t0, zero, 33
	sb t0, 2(sp)
	addi t0, zero, 10
	sb t0, 3(sp)
        addi t1, zero, 4
		addi a0, zero, 1
		add a1, zero, sp
		addi a2, zero, 4
		addi a7, zero, 64
		ecall
		addi t1, t1, -1
	bne t1, zero, ff4
	lui t2, 44434
	addi t2, t2, 577
	sw t2, 0(sp)
		addi a0, zero, 1
		add a1, zero, sp
		addi a2, zero, 4
		addi a7, zero, 64
		ecall
	addi a0, zero, 777
	addi a7, zero, 93
	ecall
.end
