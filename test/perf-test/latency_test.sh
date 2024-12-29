#!/bin/bash 
# 测试内容：连接建立后马上断开，记录时间。结果保存在latency_ssh和latency_quicssh
ip_addr='182.92.122.11'
user='hanyh'
count=100
total_time=0
quicssh_path='/home/cjh/go/bin/quicssh'
quicssh_port=4242
res_ssh_file=latency_ssh
res_quicssh_file=latency_quicssh

# 测试之前删除存放延迟结果的文件
rm $res_ssh_file
rm $res_quicssh_file

# 对原本的ssh进行测试
for i in $(seq 1 $count); do
    # echo "Test $i"
    real_time=$( (time ssh $user@$ip_addr exit) 2>&1 | grep real | awk '{print $2}')
    # echo "Real Time: $real_time"

    # 提取分钟数和秒数
    minutes=$(echo $real_time | sed -E 's/([0-9]+)m([0-9]+(\.[0-9]+)?)s/\1/')
    seconds=$(echo $real_time | sed -E 's/([0-9]+)m([0-9]+(\.[0-9]+)?)s/\2/')

    # 将分钟转换为秒并加上秒数
    total_seconds=$(echo "$minutes * 60 + $seconds" | bc)

    # 将秒转换为毫秒
    milliseconds=$(echo "$total_seconds * 1000" | bc)

    echo $milliseconds >> $res_ssh_file

    # echo "Milliseconds: $milliseconds"

    total_time=$(echo "scale=2; ${total_time:-0} + ${milliseconds:-0}" | bc)
done

# 输出总时间和平均时间
echo "total: $total_time"
avg_time=$(echo "scale=2; $total_time / $count" | bc)
echo "avg: $avg_time"

total_time=0

# 对quicssh进行测试
for i in $(seq 1 $count); do
    # echo "Test $i"
    real_time=$( (time ssh -o ProxyCommand="$quicssh_path client --addr "%h:$quicssh_port"" $user@$ip_addr exit) 2>&1 | grep real | awk '{print $2}')
    # echo "Real Time: $real_time"

    # 提取分钟数和秒数
    minutes=$(echo $real_time | sed -E 's/([0-9]+)m([0-9]+(\.[0-9]+)?)s/\1/')
    seconds=$(echo $real_time | sed -E 's/([0-9]+)m([0-9]+(\.[0-9]+)?)s/\2/')

    # 将分钟转换为秒并加上秒数
    total_seconds=$(echo "$minutes * 60 + $seconds" | bc)

    # 将秒转换为毫秒
    milliseconds=$(echo "$total_seconds * 1000" | bc)

    echo $milliseconds >> $res_quicssh_file
    # echo "Milliseconds: $milliseconds"

    total_time=$(echo "scale=2; ${total_time:-0} + ${milliseconds:-0}" | bc)
done

# 输出总时间和平均时间
echo "total: $total_time"
avg_time=$(echo "scale=2; $total_time / $count" | bc)
echo "avg: $avg_time"