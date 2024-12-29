import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np

# 定义函数读取文件并将数保存到数组中
def read_numbers_from_file(file_path):
    numbers = []  # 用于存储数字的数组
    try:
        with open(file_path, 'r') as file:  # 以只读模式打开文件
            for line in file:
                # 去掉行首和行尾的空白字符（包括换行符），并将内容转换为数字
                number = line.strip()
                if number:  # 确保行不为空
                    numbers.append(float(number))  # 将数值存入数组（用 float 保存浮点数）
        return numbers
    except FileNotFoundError:
        print(f"文件 {file_path} 未找到！")
        return []
    except ValueError as e:
        print(f"文件内容存在无法转换为数字的项：{e}")
        return []

def draw(ssh_latency, quic_latency, filename_prefix, box_ylim_s=-1, box_ylim_e=-1):
    # 绘制箱线图
    data = [ssh_latency, quic_latency]
    labels = ['SSH', 'QUIC-SSH']

    plt.figure(figsize=(8, 6))
    sns.boxplot(data=data)
    plt.xticks([0, 1], labels)
    plt.title("SSH vs QUIC-SSH Latency Comparison" + " " + filename_prefix)
    plt.ylabel("Latency (ms)")
    plt.xlabel("Protocol")
    # plt.legend()

    output_file = filename_prefix + '_box.png'
    plt.savefig(output_file)

    if box_ylim_s > 0:
        plt.ylim(box_ylim_s, box_ylim_e)
        output_file = filename_prefix + '_box_2.png'
        plt.savefig(output_file)

    # KDE 密度曲线图（Kernel Density Estimation）
    plt.figure(figsize=(10, 6))
    sns.kdeplot(ssh_latency, label="SSH", fill=True, color="blue")
    sns.kdeplot(quic_latency, label="QUIC-SSH", fill=True, color="orange")
    plt.title("Latency Density Comparison" + " " + filename_prefix)
    plt.xlabel("Latency (ms)")
    plt.ylabel("Density")
    plt.legend()
    plt.grid(True)
    output_file = filename_prefix + '_kde.png'
    plt.savefig(output_file)
    plt.close()

    # 条形图
    # 计算平均值
    mean_ssh = np.mean(ssh_latency)
    mean_quic = np.mean(quic_latency)
    means = [mean_ssh, mean_quic]
    labels = ['SSH', 'QUIC-SSH']

    plt.figure(figsize=(8, 6))
    plt.bar(labels, means, color=['blue', 'orange'])
    plt.title("Average Latency Comparison" + " " + filename_prefix)
    plt.ylabel("Average Latency (ms)")
    output_file = filename_prefix + '_bar.png'
    plt.savefig(output_file)
    plt.close()

# 示例数据
ssh_latency = read_numbers_from_file('./latency_ssh_1KB')
quic_latency = read_numbers_from_file('./latency_quicssh_1KB')
draw(ssh_latency, quic_latency, 'latency_1KB')

ssh_latency = read_numbers_from_file('./latency_ssh_1MB')
quic_latency = read_numbers_from_file('./latency_quicssh_1MB')
draw(ssh_latency, quic_latency, 'latency_1MB', 300, 600)

ssh_latency = read_numbers_from_file('./latency_ssh_5MB')
quic_latency = read_numbers_from_file('./latency_quicssh_5MB')
draw(ssh_latency, quic_latency, 'latency_5MB')

ssh_latency = read_numbers_from_file('./latency_ssh_10MB')
quic_latency = read_numbers_from_file('./latency_quicssh_10MB')
draw(ssh_latency, quic_latency, 'latency_10MB')