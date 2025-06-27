# HTML2Image API 文档

## 概述

HTML2Image 是一个将 HTML 内容转换为图片的 Web API 服务。它使用 Chrome 无头浏览器来渲染 HTML，并返回 base64 编码的 PNG 图片。

## API 端点

### 1. 转换 HTML 到图片

**端点**: `POST /convert`

**请求体** (JSON):
```json
{
    "html": "<div>你的 HTML 内容</div>",
    "html_base64": "PGRpdj7kvaDnmoQgSFRNTCDlhoXlrrkK...",  // 可选，base64 编码的 HTML（优先使用）
    "use_local_server": false  // 可选，是否使用本地服务器方式（默认 false）
}
```

**参数说明**:
- `html`: 原始 HTML 字符串（需要正确的 JSON 转义）
- `html_base64`: Base64 编码的 HTML 字符串（推荐用于复杂 HTML）
- `use_local_server`: 是否使用本地服务器模式（更可靠地加载外部资源）

注意：`html` 和 `html_base64` 至少需要提供一个，如果同时提供，优先使用 `html_base64`。

**响应** (JSON):
```json
{
    "success": true,
    "data": "iVBORw0KGgoAAAANS..."  // base64 编码的 PNG 图片
}
```

**错误响应**:
```json
{
    "success": false,
    "error": "错误信息"
}
```

### 2. 健康检查

**端点**: `GET /health`

**响应**:
```json
{
    "healthy": true
}
```

### 3. 服务信息

**端点**: `GET /`

**响应**:
```json
{
    "service": "HTML2Image API",
    "version": "1.0.0",
    "endpoints": "POST /convert - 将 HTML 转换为图片"
}
```

## 使用示例

### cURL 示例

```bash
# 基本使用
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{
    "html": "<html><body><h1>Hello World</h1></body></html>"
  }' | jq -r '.data' | base64 -d > output.png

# 使用本地服务器方式（更稳定，适合加载外部资源）
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{
    "html": "<html><body><h1>Hello</h1><img src=\"https://example.com/image.png\"></body></html>",
    "use_local_server": true
  }' | jq -r '.data' | base64 -d > output.png

# 使用 base64 编码的 HTML（适合包含特殊字符的复杂 HTML）
html_content='<div class="test">包含"引号"和特殊字符</div>'
html_base64=$(echo -n "$html_content" | base64)

curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d "{
    \"html_base64\": \"$html_base64\",
    \"use_local_server\": true
  }" | jq -r '.data' | base64 -d > output.png
```

### JavaScript 示例

```javascript
async function convertHTMLToImage(html) {
    const response = await fetch('http://localhost:8080/convert', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            html: html,
            use_local_server: true  // 推荐用于加载外部图片
        })
    });

    const result = await response.json();
    
    if (result.success) {
        // 将 base64 转换为 Blob
        const byteCharacters = atob(result.data);
        const byteNumbers = new Array(byteCharacters.length);
        for (let i = 0; i < byteCharacters.length; i++) {
            byteNumbers[i] = byteCharacters.charCodeAt(i);
        }
        const byteArray = new Uint8Array(byteNumbers);
        const blob = new Blob([byteArray], {type: 'image/png'});
        
        // 创建下载链接
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'converted.png';
        a.click();
        
        // 清理
        URL.revokeObjectURL(url);
    } else {
        console.error('转换失败:', result.error);
    }
}

// 使用示例
const html = `
<div style="padding: 20px; font-family: Arial;">
    <h1>测试标题</h1>
    <p>这是一个测试段落。</p>
    <img src="https://via.placeholder.com/150" alt="测试图片">
</div>
`;
convertHTMLToImage(html);
```

### Python 示例

```python
import requests
import base64
from PIL import Image
from io import BytesIO

def convert_html_to_image(html, output_file='output.png'):
    url = 'http://localhost:8080/convert'
    
    payload = {
        'html': html,
        'use_local_server': True  # 推荐用于加载外部图片
    }
    
    response = requests.post(url, json=payload)
    result = response.json()
    
    if result['success']:
        # 解码 base64 图片
        img_data = base64.b64decode(result['data'])
        
        # 保存图片
        with open(output_file, 'wb') as f:
            f.write(img_data)
        
        # 或者使用 PIL 处理图片
        img = Image.open(BytesIO(img_data))
        img.show()
        
        print(f'图片已保存到 {output_file}')
    else:
        print(f'转换失败: {result["error"]}')

# 使用示例
html = '''
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        h1 { color: #333; }
    </style>
</head>
<body>
    <h1>Python 示例</h1>
    <p>这是通过 Python 调用 API 生成的图片。</p>
</body>
</html>
'''

convert_html_to_image(html)
```

## 注意事项

1. **图片大小限制**: 生成的图片大小取决于 HTML 内容的复杂度，建议控制在合理范围内。

2. **超时设置**: API 有 60 秒的超时限制，复杂页面可能需要更多时间。

3. **外部资源加载**: 
   - 使用 `use_local_server: true` 可以更可靠地加载外部图片和资源
   - 默认方式（data URL）更快但可能遇到 CORS 限制

4. **并发限制**: 由于使用 Chrome 渲染，建议控制并发请求数量以避免资源耗尽。

5. **字体支持**: 确保服务器安装了所需的字体，特别是中文字体。

## 部署建议

- 使用 Docker 部署以确保环境一致性
- 至少分配 2GB 内存
- 考虑使用负载均衡处理高并发
- 监控 Chrome 进程的资源使用情况 