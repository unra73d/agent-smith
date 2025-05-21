

class MCPList extends List {
    constructor() {
        super();
        this.testMCPController = null

        document.addEventListener('storage:mcps', e => this.items = e.detail);
        document.addEventListener('mcps:new', async e => { this.onNewMCP() });
        this._initStyle()
    }

    async _initStyle() {
        await super._initStyle()
        this.shadowRoot.adoptedStyleSheets = [
            ...this.shadowRoot.adoptedStyleSheets,
            await loadCSS('components/mcp/mcp.css')
        ];
    }

    async onNewMCP() {
        const existingNames = (this.items || []).map(item => item.name);

        const fields = [
            { name: 'name', label: 'MCP name', type: 'text', required: true },
            {
                name: 'type',
                label: 'Type',
                type: 'select',
                required: true,
                options: [
                    { value: 'sse', label: 'SSE' },
                    { value: 'stdio', label: 'stdio' }
                ]
            },
            {
                name: 'url',
                label: 'URL endpoint',
                type: 'text',
                required: true,
                visibleIf: { type: 'sse' }
            },
            {
                name: 'command',
                label: 'Command',
                type: 'text',
                required: true,
                visibleIf: { type: 'stdio' }
            },
            {
                name: 'args',
                label: 'Command arguments',
                type: 'text',
                required: false,
                visibleIf: { type: 'stdio' }
            }
        ];

        const validate = values => {
            if (existingNames.includes(values.name)) {
                return 'MCP name must be unique';
            }
        };

        const buttons = [
            {
                name: 'Test MCP',
                onClick: async (values, dialog, setStatus) => {
                    if (this.testMCPController) {
                        this.testMCPController.abort()
                        this.testMCPController = null
                    }
                    this.testMCPController = new AbortController();

                    setStatus('Testing MCP...', false);
                    try {
                        const ok = await apiMCPTest(values, this.testMCPController.signal);
                        if (ok != 'canceled') {
                            if (ok) {
                                setStatus('MCP test successful!', false);
                            } else {
                                setStatus('MCP test failed.', true);
                            }
                        }
                    } catch (err) {
                        setStatus('Error testing MCP: ' + (err.message || err), true);
                    }
                }
            }
        ];

        let res = await showEditDialog({
            title: 'New MCP server',
            fields,
            validate,
            buttons,
            onClose: () => {
                if (this.testMCPController) {
                    this.testMCPController.abort()
                    this.testMCPController = null
                }
            }
        });

        if (res) {
            await apiMCPCreate(res);
        }
    }

    getItem(data) {
        const item = document.createElement('div')

        item.innerHTML = `
            <div class="item-header">
                <ui-checkbox class="select-all-checkbox" label="${data.name}" ${data.active ? 'checked' : ''}></ui-checkbox>
                <div alt="Delete" class="delete-icon img-button" data-id="${data.id}">&#xe053;</div>
            </div>
            <div class="item-content"></div>
        `;

        const selectAllCheckbox = item.querySelector('ui-checkbox');

        const itemContent = item.querySelector('.item-content');
        if (!data.active) itemContent.classList.add('disabled')
        for (let tool of data.tools) {
            const toolItem = document.createElement('div');
            toolItem.classList.add('tool-item');
            toolItem.innerHTML = `
                <label class="tool-checkbox-area">
                    <ui-checkbox ${data.active ? '' : 'disabled'} ${tool.active ? 'checked' : ''}></ui-checkbox>
                </label>
                <span class="item-text">${tool.name} - ${tool.description}</span>
            `;
            itemContent.appendChild(toolItem);

            const toolCheckbox = toolItem.querySelector('ui-checkbox');
            toolCheckbox.addEventListener('change', (e) => {
                tool.active = e.target.checked; // Update the tool's active state
            });
        }

        selectAllCheckbox.addEventListener('change', (e) => {
            data.active = e.target.checked
        });

        const deleteIcon = item.querySelector('.delete-icon')

        deleteIcon.addEventListener('click', e => this.handleDeleteItem(data.id))

        return item;
    }

    async handleDeleteItem(itemId) {
        const confirmed = await confirmDialog('Are you sure you want to delete this item? This action cannot be undone.');

        if (confirmed) {
            apiMCPDelete(itemId)
        }
    }
}

customElements.define('mcp-list', MCPList);