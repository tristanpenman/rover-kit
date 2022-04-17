const maxMessages = 16;

const buttons = document.createElement('div');
buttons.id = 'buttons';
document.body.appendChild(buttons);

const heading = document.createElement('h3');
heading.innerText = `Last ${maxMessages} events`;
document.body.appendChild(heading);

const messages = document.createElement('div');
messages.id = 'messages';
document.body.appendChild(messages);

let ws;

const addButton = (text, command) => {
  const button = document.createElement('button');
  buttons.appendChild(button);
  button.innerText = text;
  button.ontouchstart = button.onmousedown = (event) => {
    ws.send(command);
  };
  button.ontouchend = button.onmouseup = (event) => {
    ws.send('stop');
  };
};

const addMessage = (text) => {
  const message = document.createElement('div');
  const content = document.createTextNode(text);
  message.classList = 'message';
  message.appendChild(content);
  messages.prepend(message);

  const count = messages.getElementsByTagName("div").length;
  if (count > maxMessages) {
    messages.removeChild(messages.lastChild);
  }
};

const attemptConnect = () => {
  ws = new WebSocket("ws://" + location.host + "/ws");

  ws.onclose = (event) => {
    addMessage('close');
    ws.close();
    addMessage('try again in 2s...');
    setTimeout(() => {
      attemptConnect();
    }, 2000);
  }

  ws.onerror = (event) => {
    console.error('error', {
      event
    });

    addMessage('error');
  };

  ws.onopen = (event) => {
    addMessage('open');
  };

  ws.onmessage = (event) => {
    addMessage('message: ' + event.data);
  };
};

attemptConnect();

addButton('Forwards', 'forwards');
addButton('Backwards', 'backwards');
addButton('Spin Clockwise', 'spin_cw');
addButton('Spin Counter-clockwise', 'spin_ccw');

for (i = 0; i < maxMessages; i++) {
  addMessage('');
}
