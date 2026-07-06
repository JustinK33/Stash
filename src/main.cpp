#include <QApplication>

#include "MainWindow.h"

int main(int argc, char *argv[]) {
  QApplication app(argc, argv);
  app.setApplicationName("Stash");
  app.setOrganizationName("Stash");
  app.setOrganizationDomain("stash.local");
  app.setQuitOnLastWindowClosed(false);

  MainWindow window;
  window.show();

  return app.exec();
}
