# Support short-hand rpmbuild invocation.
%define _sourcedir .
%define _builddir .
%define _srcrpmdir .
%define _rpmdir .

# Disable rpm magic that creates junk files.
%define __debug_install_post %{nil}
%define __os_install_post %{nil}

# Build rpms compatible with RHEL5
%define _source_filedigest_algorithm md5
%define _binary_filedigest_algorithm md5
%define _source_payload w9.gzdio
%define _binary_payload w9.gzdio

Name: tsp
Version: %{version}
Release: 1
Summary: Time series forwarder
Group: Systems
License: Proprietary
Packager: Jacek Masiulaniec <jacek.masiulaniec@gmail.com>
BuildRoot: /var/tmp/%{name}-buildroot
Conflicts: kernel = 1:2.6.18-194.32.1.el5
Conflicts: kernel = 1:2.6.18-238.5.1.el5
Conflicts: kernel-xen = 1:2.6.18-238.12.1.el5
Conflicts: kernel-xen = 1:2.6.18-274.7.1.el5

%description

%prep

%build
make -C ../../cmd/tsp-forwarder
make -C ../../cmd/tsp-poller
make -C ../../cmd/tsp-aggregator
make -C ../../cmd/tsp-controller

%install
rm -fr $RPM_BUILD_ROOT
make -C ../../cmd/tsp-forwarder install DESTDIR=$RPM_BUILD_ROOT
make -C ../../cmd/tsp-poller install DESTDIR=$RPM_BUILD_ROOT
make -C ../../cmd/tsp-aggregator install DESTDIR=$RPM_BUILD_ROOT
make -C ../../cmd/tsp-controller install DESTDIR=$RPM_BUILD_ROOT

%clean
rm -rf $RPM_BUILD_ROOT

%pre
/sbin/service tsp stop >/dev/null 2>&1 || true

%post
/sbin/chkconfig --add tsp

%preun
s="tsp"
/sbin/service $s stop >/dev/null
if [ "$1" = 0 ]; then
	/sbin/chkconfig --del $s || true
fi

%files
%defattr(644,root,root,755)
%attr(0755,root,root) /etc/init.d/tsp
%config(noreplace) /etc/sysconfig/tsp
%config(noreplace) /etc/logrotate.d/tsp
%dir /etc/tsp
%dir /etc/tsp/collect.d
%attr(0755,root,root) /usr/bin/tsp-forwarder
/usr/share/man/man*/tsp-forwarder.*
%dir /var/log/tsp

%package aggregator
Summary: Time series aggregator
Group: Systems

%description aggregator

%pre aggregator
/sbin/service tsp-aggregator stop >/dev/null 2>&1 || true

%post aggregator
/sbin/chkconfig --add tsp-aggregator

%preun aggregator
s="tsp-aggregator"
/sbin/service $s stop >/dev/null
if [ "$1" = 0 ]; then
	/sbin/chkconfig --del $s || true
fi

%files aggregator
%defattr(644,root,root,755)
%attr(0755,root,root) /etc/init.d/tsp-aggregator
%config(noreplace) /etc/sysconfig/tsp-aggregator
%config(noreplace) /etc/logrotate.d/tsp-aggregator
%dir /etc/tsp-aggregator
%attr(0755,root,root) /usr/bin/tsp-aggregator
/usr/share/man/man*/tsp-aggregator.*

%package poller
Summary: Time series poller
Group: Systems

%description poller

%pre poller
/sbin/service tsp-poller stop >/dev/null 2>&1 || true

%post poller
/sbin/chkconfig --add tsp-poller

%preun poller
s="tsp-poller"
/sbin/service $s stop >/dev/null
if [ "$1" = 0 ]; then
	/sbin/chkconfig --del $s || true
fi

%files poller
%defattr(644,root,root,755)
%attr(0755,root,root) /etc/init.d/tsp-poller
%config(noreplace) /etc/sysconfig/tsp-poller
%config(noreplace) /etc/logrotate.d/tsp-poller
%dir /etc/tsp-poller
%dir /etc/tsp-poller/collect.d
%attr(0755,root,root) /usr/bin/tsp-poller
/usr/share/man/man*/tsp-poller.*

%package controller
Summary: Time Series pipeline controller
Group: Systems

%description controller

%pre controller
/sbin/service tsp-controller stop >/dev/null 2>&1 || true

%post controller
/sbin/chkconfig --add tsp-controller

%preun controller
s="tsp-controller"
/sbin/service $s stop >/dev/null
if [ "$1" = 0 ]; then
	/sbin/chkconfig --del $s || true
fi

%files controller
%defattr(644,root,root,755)
%config(noreplace) /etc/sysconfig/*
%attr(0755,root,root) /etc/init.d/*
# %config(noreplace) /etc/logrotate.d/*
%dir /etc/tsp-controller
%config(noreplace) /etc/tsp-controller/config
%attr(0755,root,root) /usr/bin/*
/usr/share/man/man*/*
