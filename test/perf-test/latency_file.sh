#!/bin/bash 
# 测试内容：发送文件，大小分别为1KB*100次和1MB*100次和5MB*100次和10MB*100次
ip_addr='182.92.122.11'
user='hanyh'
quicssh_path='/home/cjh/go/bin/quicssh'
quicssh_port=4242

random_file_path_prefix="./random_files/"
res_file_path_prefix="./network_latency_file/"

file1="random_1KB.txt"
file2="random_1MB.txt"
file3="random_5MB.txt"
file4="random_10MB.txt"
file5="random_500MB.txt"

file1_count=100
file2_count=100
file3_count=100
file4_count=100
file5_count=1

file1_ssh_res=latency_ssh_1KB
file2_ssh_res=latency_ssh_1MB
file3_ssh_res=latency_ssh_5MB
file4_ssh_res=latency_ssh_10MB
file5_ssh_res=latency_ssh_500MB
file1_quicssh_res=latency_quicssh_1KB
file2_quicssh_res=latency_quicssh_1MB
file3_quicssh_res=latency_quicssh_5MB
file4_quicssh_res=latency_quicssh_10MB
file5_quicssh_res=latency_quicssh_500MB

# 指令一个shell命令并记录时间
command_exec_timing() {
    local command=$1
    real_time=$( (time eval $command;) 2>&1 | grep real | awk '{print $2}')
    # 提取分钟数和秒数
    minutes=$(echo $real_time | sed -E 's/([0-9]+)m([0-9]+(\.[0-9]+)?)s/\1/')
    seconds=$(echo $real_time | sed -E 's/([0-9]+)m([0-9]+(\.[0-9]+)?)s/\2/')

    # 将分钟转换为秒并加上秒数
    total_seconds=$(echo "$minutes * 60 + $seconds" | bc)

    # 将秒转换为毫秒
    milliseconds=$(echo "$total_seconds * 1000" | bc)
    echo $milliseconds
}

# param1: 要执行的命令
# param2: 执行次数
# param3: 讲结果写入哪个文件
loop_and_record() {
    local command=$1
    local count=$2
    local res_file=$3
    # echo $command
    # echo $count
    # echo $res_file

    for i in $(seq 1 $count); do
        # command_exec_timing "$command"
        milliseconds=$(command_exec_timing "$command")
        echo $milliseconds >> $res_file
    done
}

begin=5
end=5
for i in $(seq $begin $end); do
    _i=$i
    # ssh
    var1="file${i}_ssh_res"
    rm ${res_file_path_prefix}${!var1}
    var2="file${_i}"
    command="scp ${random_file_path_prefix}${!var2} $user@$ip_addr:/$user"
    var3="file${_i}_count"
    loop_and_record "$command" ${!var3} $res_file_path_prefix${!var1}

    # quicssh
    var1="file${_i}_quicssh_res"
    echo ${var1}
    echo ${res_file_path_prefix}${!var1}
    rm ${res_file_path_prefix}${!var1}
    command="scp -o ProxyCommand=\"$quicssh_path client --addr %h:$quicssh_port\" $random_file_path_prefix${!var2} $user@$ip_addr:/$user"
    loop_and_record "$command" ${!var3} $res_file_path_prefix${!var1}
done
# command_exec_timing "ssh hanyh@182.92.122.11"
# milliseconds=$(command_exec_timing "ssh hanyh@182.92.122.11")
# echo $milliseconds
