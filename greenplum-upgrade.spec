Name: greenplum-upgrade
Version: %{gpupgrade_version}
# Release is a way of versioning the spec file.
# Only bump the Release if shipping gpupgrade without also bumping the
# gpugprade_version (ie: VERSION).
Release: %{gpupgrade_rpm_release}%{?dist}
Summary: %{summary}
License: %{license}
URL: https://github.com/greenplum-db/gpupgrade
Source0: %{name}-%{version}.tar.gz
Prefix: /usr/local

%description
gpupgrade can do in-place upgrades without the need for additional hardware, disk space, and with less downtime.

%prep
# If the gpupgrade_version macro is not defined, it gets interpreted as a literal string, need %% to escape it
if [ %{gpupgrade_version} = '%%{gpupgrade_version}' ] ; then
  echo "The macro (variable) gpupgrade_version must be supplied as rpmbuild ... --define='gpupgrade_version [VERSION]'"
  exit 1
fi

%setup -q -c -n %{name}-%{version}

%install
mkdir -p %{buildroot}/%{prefix}/%{name}
cp -R * %{buildroot}/%{prefix}/%{name}
mv bash-completion.sh /etc/bash_completion.d

%files

%if "%{release_type}" == "Enterprise"
  %doc open_source_licenses.txt
%endif

%{prefix}/%{name}
