async function apiConnectSession(id){
    try {
        url = 'http://localhost:8008/agent/sessions/connect'
        if (id != undefined){
            urls = `http://localhost:8008/agent/sessions/connect/${id}`
        }
        const response = await fetch(url, {
            method: 'GET',
            headers: { 'Accept': 'application/json' },
        });

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();

        if (data && data.session && data.session.id) {
            return data.session
        } else {
            return null
        }
    } catch (error) {
        console.error("Failed to connect:", error);
        return null
    }
}

function apiCreateSession(){

}

async function apiDeleteSession(sessionId){
    try {
        const response = await fetch(`http://localhost:8008/agent/sessions/delete/${sessionId}`);
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();

        if (data.activeSession && data.activeSession.id && data.activeSession.messages && Array.isArray(data.activeSession.messages)){
            return data.activeSession
        } else {
            return null
        }
    } catch (error) {
        console.error("Failed to delete session:", error);
        return null
    }
}

async function apiListSessions(){
    try {
        const response = await fetch('http://localhost:8008/agent/sessions/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.sessions && Array.isArray(data.sessions)) {
            return data.sessions
        } else {
            return null
        }
    } catch (error) {
        console.error("Failed to fetch sessions:", error);
        return null
    }
}

function apiDirectChat(){

}

async function apiListModels(){
    try {
        const response = await fetch('http://localhost:8008/agent/models/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.models && data.activeModelID !== undefined) {
            return data
        } else {
            console.error("Invalid data structure received for models:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to populate model selector:", error);
        return null
    }
}