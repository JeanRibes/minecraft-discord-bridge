wschat,cerr=http.websocket("ws://chat-bridge.home.ribes.ovh/chat")
if cerr ~= nil then
    print(cerr)
    sleep(2)
end
chatbox = peripheral.wrap("top")
local function wsrcv()
    while true do
        local message,_ = wschat.receive()
        chatbox.sendMessage(message)
    end
end
local function chatrcv()
    while true do
        local _,username,message = os.pullEvent("chat")
        wschat.send("*<"..username..">* "..message)
    end
end
parallel.waitForAll(wsrcv,chatrcv)
    
