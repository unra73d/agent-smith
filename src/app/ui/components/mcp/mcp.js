class MCPList extends List{
    constructor(){
        super()

        document.addEventListener('storage:mcps', e=>this.items = e.detail)
    }

    getItem(data){
        const item = document.createElement('div')
        if(data.id){
            item.setAttribute("data-id", data.id)
        } else {
            item.setAttribute("data-id", crypto.randomUUID())
        }

        let text = ''
        if(data.name)text = data.name
        if(data.text)text = data.text
        if(data.content)text = data.content
        if(data.description)text = data.description
        if(data.summary)text = data.summary

        item.innerHTML = `
        <div class="item-header">${text}</div>
        <div class="item-content open">
        </div>
        `;
        const itemContent = item.querySelector('.item-content')
        for(let tool of data.tools){
            itemContent.innerHTML += `
            <input type="checkbox"/>
            <span class="item-text"><b>${tool.name}</b> - ${tool.description}</span>
            `
        }

        item.querySelector('.item-header').addEventListener('click', e=>item.querySelector('.item-content').classList.toggle('open'))
        return item
    }
}

customElements.define('mcp-list', MCPList)