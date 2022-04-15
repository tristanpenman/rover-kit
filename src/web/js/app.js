const messages = document.createElement('ul');
document.body.appendChild(messages);

const ws = new WebSocket("ws://" + location.host + "/ws");

ws.onerror = (event) => {
  console.error('error', {
    event
  });

  const message = document.createElement('li');
  const content = document.createTextNode("error: " + event);
  message.appendChild(content);
  messages.appendChild(message);
};

ws.onopen = (event) => {
  const message = document.createElement('li');
  const content = document.createTextNode("open");
  message.appendChild(content);
  messages.appendChild(message);
};

ws.onmessage = function (event) {
  const message = document.createElement('li');
  const content = document.createTextNode("message: " + event.data);
  message.appendChild(content);
  messages.appendChild(message);
};
