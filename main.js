var app = require('app');
var BrowserWindow = require('browser-window');
require('crash-reporter').start();
var mainWindow = null;
var subgo = require('child_process').spawn(__dirname+'/fsts');

const electron = require('electron');
const globalShortcut = electron.globalShortcut;


app.on('window-all-closed', function() {
  app.quit();
});


app.on('ready', function() {
    mainWindow = new BrowserWindow({width: 750,
                                    height: 399,
                                    darkTheme:true,
                                    frame: false,
                                    autosize: 1,
                                    x: 0,
                                    y: 0});
    mainWindow.setMenuBarVisibility(false);
    mainWindow.setTitle("Fsto.tv");
    mainWindow.setAlwaysOnTop(false);
    //mainWindow.openDevTools();
    mainWindow.loadUrl('http://localhost:8000?autoplay=1');

  // Emitted when the window is closed.
  var ret = globalShortcut.register('ctrl+w', function() {
    subgo.kill('SIGINT');
    app.quit();
  });
  if (!ret) {
    console.log('registration failed');
  }

  var h = globalShortcut.register('ctrl+`', function() {
      if (mainWindow.isVisible()){mainWindow.hide();} else {mainWindow.show();}
  });

  // Check whether a shortcut is registered.
  console.log(globalShortcut.isRegistered('ctrl+`'));
});

app.on('will-quit', function() {
  // Unregister a shortcut.
  globalShortcut.unregister('ctrl+w');
  globalShortcut.unregister('ctrl+`');

  // Unregister all shortcuts.
  globalShortcut.unregisterAll();
  mainWindow.on('closed', function() {
    mainWindow = null;
    if (subgo){subgo.kill('SIGINT')}
    app.quit();
  });
});
