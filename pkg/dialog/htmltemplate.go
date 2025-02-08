package dialog

const dialogTreeHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>上下文🌳管理</title>
    <style>
        /* 基础容器 */
        .dialog-container {
            max-width: 100%;
            margin: 20px;
            padding: 0 20px;
        }

        /* 分组样式 */
        .group {
            margin: 25px 0;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            background: white;
        }
        .group-header {
            padding: 16px 20px;
            background: #f8f9fa;
            border-radius: 8px 8px 0 0;
            cursor: pointer;
            display: flex;
            align-items: center;
            border-bottom: 2px solid #dee2e6;
        }
        .group-title {
            font-size: 1.1em;
            font-weight: 600;
            color: #2b2d42;
        }

        /* 对话树样式 */
        .dialog-tree {
            padding: 15px 20px;
        }
        .node {
            margin: 12px 0;
            position: relative;
        }

        /* 角色标签样式 */
        .role-tag {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 4px;
            margin-right: 15px;
            font-size: 0.85em;
            font-weight: 500;
            min-width: 80px;
            text-align: center;
        }

        /* 角色颜色方案 */
        .role-user .role-tag {
            background: #E3F2FD;
            color: #0D47A1;
            border: 1px solid #90CAF9;
        }
        .role-assistant .role-tag {
            background: #E8F5E9;
            color: #1B5E20;
            border: 1px solid #A5D6A7;
        }
        .role-system .role-tag {
            background: #F3E5F5;
            color: #4A148C;
            border: 1px solid #CE93D8;
        }

        /* 内容样式 */
        .content-text {
            color: #424242;
            line-height: 1.6;
        }

        /* 折叠控制 */
        .toggle {
            cursor: pointer;
            margin-right: 12px;
            width: 20px;
            text-align: center;
            color: #757575;
        }
        .children {
            padding-left: 40px;
            border-left: 2px solid #eee;
            margin: 8px 0;
        }
        .collapsed .children {
            display: none;
        }
    </style>
</head>
<body>
    <h1>{{.Welcome}}</h1>
    <div class="dialog-container">
        {{range $groupKey, $trees := .IndexedDialogTreesmap}}
        <div class="group">
            <div class="group-header" onclick="toggleGroup(this)">
                <span class="toggle">▶</span>
                <span class="group-title">{{$groupKey}}</span>
            </div>
            <div class="dialog-tree">
                {{range $trees}}
                <div class="node">
                    <div class="role-{{.Role}}">
                        <span class="toggle" onclick="toggleNode(this)">▶</span>
                        <span class="role-tag">{{.Role}}</span>
                        <span class="content-text">{{.Content}}</span>
                    </div>
                    {{if .Children}}
                    <div class="children">
                        {{template "childNodes" .Children}}
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
    </div>

    <script>
        // 分组切换
        function toggleGroup(header) {
            const content = header.parentNode.querySelector('.dialog-tree')
            const toggle = header.querySelector('.toggle')
            content.style.display = content.style.display === 'none' ? 'block' : 'none'
            toggle.textContent = content.style.display === 'none' ? '▶' : '▼'
        }

        // 节点切换
        function toggleNode(toggle) {
            const node = toggle.closest('.node')
            const children = node.querySelector('.children')
            if (children) {
                children.style.display = children.style.display === 'none' ? 'block' : 'none'
                toggle.textContent = children.style.display === 'none' ? '▶' : '▼'
            }
        }

        // 初始状态：默认折叠所有分组和子节点
        document.querySelectorAll('.dialog-tree').forEach(t => t.style.display = 'none')
        document.querySelectorAll('.children').forEach(c => c.style.display = 'none')
    </script>
</body>
</html>

{{define "childNodes"}}
    {{range .}}
    <div class="node">
        <div class="role-{{.Role}}">
            {{if .Children}}<span class="toggle" onclick="toggleNode(this)">▶</span>{{end}}
			<span class="role-tag">
                <a href="/{{.Id}}/{{.MessageId}}">{{.Role}}</a>
			</span>
            <span class="content-text">{{.Content}}</span>
        </div>
        {{if .Children}}
        <div class="children">
            {{template "childNodes" .Children}}
        </div>
        {{end}}
    </div>
    {{end}}
{{end}}
`

const dialogListHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>上下文列表</title>
    <!-- 引入 marked.js -->
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <!-- 引入 highlight.js -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/github.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/highlight.min.js"></script>
    <!-- 引入 MathJax -->
    <script src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"></script>
    <style>
        /* 保持原有样式，增加 Markdown 元素适配 */
        body { font-family: -apple-system, sans-serif; background: #f8f9fa; }
        .chat-container { max-width: 800px; margin: 20px auto; background: white; border-radius: 12px; box-shadow: 0 2px 12px rgba(0,0,0,0.1); padding: 24px; }
        .message { margin: 16px 0; }
        .user-message { background: #007bff; color: white; border-radius: 15px 15px 0 15px; padding: 12px 16px; max-width: 70%; margin-left: auto; }
        .assistant-message { background: #e9ecef; color: #212529; border-radius: 15px 15px 15px 0; padding: 12px 16px; max-width: 70%; }
        .role-label { font-size: 0.85em; color: #6c757d; margin-bottom: 4px; }

        /* Markdown 元素样式 */
        .message-content strong { font-weight: 600; }
        .message-content code { background: rgba(175,184,193,0.2); padding: 0.2em 0.4em; border-radius: 4px; }
        .message-content pre { background: #f6f8fa; padding: 16px; border-radius: 6px; overflow-x: auto; }
        .message-content pre code { background: transparent; padding: 0; display: block; }
        .message-content a { color: #007bff; text-decoration: none; }
        .message-content a:hover { text-decoration: underline; }
        .message-content ul { padding-left: 20px; }
        .message-content li { margin: 4px 0; }
        /* LaTeX 公式样式 */
        .message-content .mathjax { font-size: 1.1em; }
    </style>
</head>
<body>
    <div class="chat-container">
        {{range .}}
        <div class="message">
            <div class="role-label">
                {{if eq .Role "user"}}你{{else}}助手{{end}}
            </div>
            <div class="{{if eq .Role "user"}}user-message{{else}}assistant-message{{end}}">
                <!-- 原始 Markdown 内容存放在隐藏的 pre 标签中 -->
                <pre class="raw-markdown" style="display: none;">{{.Content}}</pre>
                <!-- 渲染后的内容显示在这里 -->
                <div class="message-content"></div>
            </div>
        </div>
        {{end}}
    </div>

    <script>
        // 配置 marked
        marked.setOptions({
            breaks: true,    // 自动换行
            highlight: function(code, lang) {
                // 使用 highlight.js 进行代码高亮
                const language = hljs.getLanguage(lang) ? lang : 'plaintext';
                return hljs.highlight(code, { language }).value;
            }
        });

        // 渲染所有 Markdown 内容
        document.querySelectorAll('.raw-markdown').forEach(pre => {
            const container = pre.nextElementSibling;
            const rawMarkdown = pre.textContent;
            
            // 渲染 Markdown
            container.innerHTML = marked.parse(rawMarkdown);
            
            // 移除原始内容
            pre.remove();
        });

        // 自动滚动到底部
        window.scrollTo(0, document.body.scrollHeight);

        // 配置 MathJax
        MathJax = {
            tex: {
                inlineMath: [['$', '$'], ['\\(', '\\)']], // 行内公式分隔符
                displayMath: [['$$', '$$'], ['\\[', '\\]']], // 块级公式分隔符
                processEscapes: true, // 允许使用 \ 转义
            },
            options: {
                skipHtmlTags: ['script', 'noscript', 'style', 'textarea', 'pre'], // 跳过指定标签
            },
            startup: {
                pageReady: () => {
                    // 页面加载完成后渲染公式
                    return MathJax.startup.defaultPageReady();
                }
            }
        };
    </script>
</body>
</html>
`
