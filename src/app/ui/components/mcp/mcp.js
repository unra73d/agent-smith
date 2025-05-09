class MCPList extends List{
    constructor(){
        super()

        document.addEventListener('storage:mcps', e=>this.items = e.detail)
    }
}

customElements.define('mcp-list', MCPList)