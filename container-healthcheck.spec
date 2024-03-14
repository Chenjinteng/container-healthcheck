BuildArch:     x86_64
Name:          container-healthcheck
Version:       1.0.0
Release:       2.el7
License:       BSD
Group:         default
Summary:       A Container Health Check Services
URL:           http://www.sretalk.com/
Packager:      sretalker
BuildRoot:     /data/rpmbuild/

%description
A Container Health Check Services.

%pre
# test ! -f /var/log/container-healthcheck/ || \
#   mkdir -p /var/log/container-healthcheck/

%install
\cp -rfp /data/rpmbuild/SOURCES/container-healthcheck-V1.0.0-release/* ${RPM_BUILD_ROOT}
chmod +x ${RPM_BUILD_ROOT}/usr/local/bin/container-healthcheck

%files
%defattr (-,root,root,0755)
/var/log/container-healthcheck
/etc/rsyslog.d/container-healthcheck.conf
/etc/sysconfig/container-healthcheck
/usr/local/bin/container-healthcheck
/usr/lib/systemd/system/container-healthcheck.service


%post -p /bin/sh
#!/bin/sh

if command -v systemctl >/dev/null; then
    systemctl daemon-reload || true
fi
exit 0

%changelog