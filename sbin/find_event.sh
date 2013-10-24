#!/bin/bash
SED_OPTS='-u'
if [[ $(uname) == "Darwin" ]]; then
  SED_OPTS='-l'
fi
PATTERN="event_id\":\"$2"
if [[ $(tail -n1 $1) == *"$PATTERN"* ]]
then
  echo -n "-1"
else
  sed $SED_OPTS "1,/\"$PATTERN\"/d" $1 | wc -l | tr -d ' \n'
fi
