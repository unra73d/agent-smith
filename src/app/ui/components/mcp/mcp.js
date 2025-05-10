class MCPList extends List {
    constructor() {
        super();

        const styles = document.createElement('style');
        styles.innerHTML = `
        @import url('global.css');
        @import url('components/mcp/mcp.css');
        `
        this.shadowRoot.appendChild(styles);

        document.addEventListener('storage:mcps', e => this.items = e.detail);
    }

    getItem(data) {
        const item = document.createElement('div')

        item.innerHTML = `
            <div class="item-header">
                <input type="checkbox" class="select-all-checkbox">
                <span class="header-text">${data.name}</span>
                <img src="icons/delete.svg" alt="Delete" class="delete-icon" data-id="${data.id}">
            </div>
            <div class="item-content open">
            </div>
        `;

        const itemContent = item.querySelector('.item-content');
        for (let tool of data.tools) {
            itemContent.innerHTML += `
                <div class="tool-item">
                    <input type="checkbox"/>
                    <span class="item-text"><b>${tool.name}</b> - ${tool.description}</span>
                </div>
            `;
        }

        const selectAllCheckbox = item.querySelector('.select-all-checkbox');
        selectAllCheckbox.addEventListener('change', e => {
            const isChecked = e.target.checked;
            itemContent.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
                checkbox.checked = isChecked;
            });
        });

        const deleteIcon = item.querySelector('.delete-icon')
        item.querySelector('.item-header').addEventListener('click', e => {
            if (e.target !== selectAllCheckbox && e.target !== deleteIcon) {
                itemContent.classList.toggle('open');
            }
        });

        deleteIcon.addEventListener('click', e => this.handleDeleteItem(e, item, data.id))

        return item;
    }

    async handleDeleteItem(e, item, itemId) {
        e.stopPropagation();
        const confirmed = await confirmDialog('Are you sure you want to delete this item? This action cannot be undone.');

        if (confirmed) {

        }
    }
}

customElements.define('mcp-list', MCPList);