/******************************************************************************
 * Filename:        ts_length.c  
 * Created on:      Mar 8, 2010  
 * Author:          jeremiah  
 * Description:     打印TS各PID的时间长度  
 *  
 ******************************************************************************/ 
 
#include <stdio.h>  
#include <stdint.h>  
 
#define TS_SYNC_BYTE 0x47  
#define TS_PACKET_SIZE 188  
 
typedef struct {  
    unsigned pid;  
    double clock_begin;  
    double clock_end;  
} pid_t;  
 
pid_t pid_array[8191]; // 裤子说一个ts最多有8191个pid。那就建立一个8191的数组。  
unsigned char buf[TS_PACKET_SIZE];  
 
void get_length(unsigned char* pkt);  
void store_pid(unsigned pid, double clock);  
 
int main(int argc, char **argv) {  
    if (argc < 2) {  
        fprintf(stderr, "please use %s <file_name>/n", argv[0]);  
        return 1;  
    }  
 
    FILE *fp = fopen(argv[1], "rb");  
    if (!fp) {  
        perror("fopen");  
        return 1;  
    }  
    fseek(fp, 0, SEEK_END);  
    int size = ftell(fp);  
    rewind(fp);  
 	printf("file size %v \n",size);
    while (size > 0) {  
        int read_size = fread(buf, 1, sizeof(buf), fp);  
        /*每次读取188*/
        printf("read %d every time \n",read_size);
        size -= read_size;  
        get_length(buf);  
    }  
 
    int i;  
    for (i = 0; i < 8191; i++) {  
        if (pid_array[i].pid != 0) {  
            printf("PID:0x%x length:%fs/n", pid_array[i].pid,   
                pid_array[i].clock_end - pid_array[i].clock_begin);  
        } else {  
            break;  
        }  
    }  
    return 0;  
}  
 
void get_length(unsigned char* pkt) {  
    // Sanity check: Make sure we start with the sync byte:  
    if (pkt[0] != TS_SYNC_BYTE) {  
        fprintf(stderr, "Missing sync byte! \n");  
        return;  
    }  
 
    // If this packet doesn't contain a PCR, then we're not interested in it:  
    uint8_t const adaptation_field_control = (pkt[3] & 0x30) >> 4;  
    if (adaptation_field_control != 2 && adaptation_field_control != 3) {  
    	printf("this packet adaptation_field_control err  \n");
        return;  
    }  
 
    // there's no adaptation_field  
    uint8_t const adaptation_field_length = pkt[4];  
    if (adaptation_field_length == 0) {  
    	printf("this packet has no adaptation_field_length  \n");
        return;  
    }  
 
    // no PCR  
    uint8_t const pcr_flag = pkt[5] & 0x10;  
    if (pcr_flag == 0) {  
 	    printf("this packet has no pcr  \n");
        return;  
    }  
 
    // yes, we get a pcr  
    uint32_t pcr_base_high = (pkt[6] << 24) | (pkt[7] << 16) | (pkt[8] << 8)  
                        | pkt[9];  
    printf("yes ,we got a pcr %d",pcr_base_high);
    // caculate the clock  
    double clock = pcr_base_high / 45000.0;  
    if ((pkt[10] & 0x80)) {  
        clock += 1 / 90000.0; // add in low-bit (if set)  
    }  
    unsigned short pcr_extra = ((pkt[10] & 0x01) << 8) | pkt[11];  
    clock += pcr_extra / 27000000.0;  
 
    unsigned pid = ((pkt[1] & 0x1F) << 8) | pkt[2];  
    printf("we got a pid %ld and clock %d \n",pid,clock);
    store_pid(pid, clock);  
}  
 
void store_pid(unsigned pid, double clock) {  
    int i;  
    for (i = 0; i < 8191; i++) {  
        if (pid == pid_array[i].pid) {  
            break;  
        }  
    }  
    if (i == 8191) {  
        for (i = 0; i < 8191; i++) {  
            if (pid_array[i].pid == 0) {  
                break;  
            }  
        }  
        pid_array[i].pid = pid;  
        pid_array[i].clock_begin = clock;  
    } else {  
        pid_array[i].clock_end = clock;  
    }  
} 