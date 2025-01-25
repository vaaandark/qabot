package dialog

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>上下文管理</title>
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
{{end}}
`
