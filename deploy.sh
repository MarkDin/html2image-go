#!/bin/bash

# HTML2Image-Go Docker 部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
IMAGE_NAME="html2image-go"
IMAGE_TAG="latest"
CONTAINER_NAME="html2image-service"

echo -e "${GREEN}=== HTML2Image-Go Docker 部署脚本 ===${NC}"

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker 未安装${NC}"
    exit 1
fi

# 构建 Docker 镜像
echo -e "${YELLOW}正在构建 Docker 镜像...${NC}"
docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}镜像构建成功！${NC}"
else
    echo -e "${RED}镜像构建失败！${NC}"
    exit 1
fi

# 停止并删除旧容器（如果存在）
if docker ps -a | grep -q ${CONTAINER_NAME}; then
    echo -e "${YELLOW}停止并删除旧容器...${NC}"
    docker stop ${CONTAINER_NAME} 2>/dev/null || true
    docker rm ${CONTAINER_NAME} 2>/dev/null || true
fi

# 运行新容器
echo -e "${YELLOW}启动新容器...${NC}"
docker run -d \
    --name ${CONTAINER_NAME} \
    --restart unless-stopped \
    -p 8080:8080 \
    -v $(pwd)/output:/app/output \
    --cap-add SYS_ADMIN \
    ${IMAGE_NAME}:${IMAGE_TAG}

if [ $? -eq 0 ]; then
    echo -e "${GREEN}容器启动成功！${NC}"
    echo -e "${GREEN}容器名称: ${CONTAINER_NAME}${NC}"
    echo -e "${GREEN}输出目录: $(pwd)/output${NC}"
    
    # 显示容器日志
    echo -e "\n${YELLOW}容器日志:${NC}"
    docker logs ${CONTAINER_NAME}
    
    # 测试运行
    echo -e "\n${YELLOW}等待容器完全启动...${NC}"
    sleep 3
    
    # 检查输出文件
    if [ -f "output/output.png" ]; then
        echo -e "${GREEN}测试成功！输出文件已生成: output/output.png${NC}"
    else
        echo -e "${YELLOW}提示: 输出文件可能需要更多时间生成${NC}"
    fi
else
    echo -e "${RED}容器启动失败！${NC}"
    exit 1
fi

echo -e "\n${GREEN}=== 部署完成 ===${NC}"
echo -e "${YELLOW}查看日志: docker logs -f ${CONTAINER_NAME}${NC}"
echo -e "${YELLOW}停止服务: docker stop ${CONTAINER_NAME}${NC}"
echo -e "${YELLOW}启动服务: docker start ${CONTAINER_NAME}${NC}" 