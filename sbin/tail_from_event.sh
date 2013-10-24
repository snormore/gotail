#!/bin/bash
SED_OPTS='-u'
if [[ $(uname) == "Darwin" ]]; then
  SED_OPTS='-l'
fi
TAIL_OPTS=""
if [[ $1 == "true" ]]; then
  TAIL_OPTS="-f"
fi
TAIL_OPTS+=" -n +0"
tail $TAIL_OPTS $2 | sed $SED_OPTS "1,/\"event_id\":\"$3\"/d"