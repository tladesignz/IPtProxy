#!/bin/sh

prepare()
{
  artifact=$1
  version=$2

  cd "$(dirname "$0")" || exit 1
  mkdir -p "bundle"
  rm -rf bundle/**
  cp -af IPtProxy.aar bundle/${artifact}-${version}.aar
  cp -af IPtProxy-sources.jar bundle/${artifact}-${version}-sources.jar
}

pom()
{
  artifact=$1
  version=$2

  cat > bundle/${artifact}-${version}.pom <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <packaging>aar</packaging>
  <groupId>com.netzarchitekten</groupId>
  <artifactId>${artifact}</artifactId>
  <version>${version}</version>
  <name>IPtProxy</name>
  <description>Obfs4proxy/Lyrebird and Snowflake Pluggable Transports for iOS, MacOS and Android</description>
  <url>https://github.com/tladesignz/IPtProxy</url>
  <inceptionYear>2020</inceptionYear>
  <licenses>
    <license>
      <name>MIT</name>
      <url>https://github.com/tladesignz/IPtProxy/blob/master/LICENSE</url>
      <distribution>repo</distribution>
    </license>
    <license>
      <name>GPL3</name>
      <url>https://gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/-/raw/main/LICENSE-GPL3.txt</url>
      <distribution>repo</distribution>
    </license>
    <license>
      <name>BSD-3-clause</name>
      <url>https://gitweb.torproject.org/pluggable-transports/snowflake.git/tree/LICENSE</url>
      <distribution>repo</distribution>
    </license>
  </licenses>
  <developers>
    <developer>
      <id>guardianproject</id>
      <name>Guardian Project</name>
      <email>support@guardianproject.info</email>
    </developer>
    <developer>
      <id>torproject</id>
      <name>Tor Project</name>
      <email>torbrowser@torproject.org</email>
    </developer>
  </developers>
  <scm>
    <connection>scm:git:https://github.com/tladesignz/IPtProxy.git</connection>
    <url>https://github.com/tladesignz/IPtProxy</url>
  </scm>
  <issueManagement>
    <url>https://github.com/tladesignz/IPtProxy/issues</url>
    <system>GitHub</system>
  </issueManagement>
</project>
EOF
}

sign()
{
  artifact=$1
  version=$2
  forceKey=""

  if [ ! -z "$3" ]; then
    forceKey="--local-user $3"
  fi

  if gpg --list-secret-keys | grep -Eo '[0-9A-F]{40}'; then
  	for f in bundle/${artifact}-*${version}*.*; do
  	    gpg ${forceKey} --armor --detach-sign $f
  	done
  fi
}

bundle()
{
  artifact=$1
  version=$2
  file=bundle-${artifact}-${version}.jar

  rm -f $file

  cd bundle || exit 1

  jar cvf ../$file ${artifact}-${version}*.*

  cd ..

  rm -r bundle
}

artifact=IPtProxy
version=$1
keyId=$2

if [ -z "$version" ]; then
  echo "Usage: ./bundle.sh <semver> [<signing key ID>]"
  exit
fi

prepare $artifact $version
pom $artifact $version
sign $artifact $version $keyId
bundle $artifact $version
