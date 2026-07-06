#include "ClipItemWidget.h"

#include <QHBoxLayout>
#include <QLabel>
#include <QPushButton>
#include <QSizePolicy>

namespace {
QPushButton *makeCopyButton(QWidget *parent) {
  auto *button = new QPushButton(QStringLiteral("Copy"), parent);
  button->setObjectName(QStringLiteral("CopyButton"));
  button->setFixedHeight(30);
  return button;
}

QPushButton *makeDeleteButton(QWidget *parent) {
  auto *button = new QPushButton(QStringLiteral("Delete"), parent);
  button->setObjectName(QStringLiteral("DeleteButton"));
  button->setFixedHeight(30);
  return button;
}
} // namespace

ClipItemWidget::ClipItemWidget(const ClipItem &clip, QWidget *parent)
    : QWidget(parent), clipText(clip.text) {
  setObjectName(QStringLiteral("ClipItem"));

  auto *layout = new QHBoxLayout(this);
  layout->setContentsMargins(12, 8, 12, 8);
  layout->setSpacing(8);

  textLabel = new QLabel(clip.text, this);
  textLabel->setObjectName(QStringLiteral("ClipText"));
  textLabel->setWordWrap(true);
  textLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Preferred);

  copyButton = makeCopyButton(this);
  deleteButton = makeDeleteButton(this);

  layout->addWidget(textLabel, 1);
  layout->addWidget(copyButton, 0, Qt::AlignTop);
  layout->addWidget(deleteButton, 0, Qt::AlignTop);

  connect(copyButton, &QPushButton::clicked, this,
          [this, clipText = this->clipText](bool) {
            emit copyClicked(clipText);
          });
  connect(deleteButton, &QPushButton::clicked, this,
          [this, clipText = this->clipText](bool) {
            emit deleteClicked(clipText);
          });
}

QString ClipItemWidget::text() const {
  return clipText;
}
