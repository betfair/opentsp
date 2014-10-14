# Support short-hand rpmbuild invocation.
%define _sourcedir .
%define _builddir .
%define _srcrpmdir .
%define _rpmdir .

# Disable rpm magic that creates junk files.
%define __debug_install_post %{nil}
%define __os_install_post %{nil}

Name: tsp-controller
Version: %{version}
Release: 1
Summary: Time Series pipeline controller
Group: Systems
License: Proprietary
Packager: Jacek Masiulaniec <jacek.masiulaniec@gmail.com>
BuildRoot: /var/tmp/%{name}-buildroot
Conflicts: kernel = 1:2.6.18-194.32.1.el5

%description

%prep

%build
cd ..
make

%install
cd ..
rm -fr $RPM_BUILD_ROOT
make install DESTDIR=$RPM_BUILD_ROOT

%clean
rm -rf $RPM_BUILD_ROOT

%post
/sbin/chkconfig --add tsp-controller

%preun
/sbin/service tsdb-controller stop >/dev/null || true
s="tsp-controller"
/sbin/service $s stop >/dev/null
if [ "$1" = 0 ]; then
	/sbin/chkconfig --del $s || true
fi

%files
%defattr(644,root,root,755)
%config(noreplace) /etc/sysconfig/*
%attr(0755,root,root) /etc/init.d/*
# %config(noreplace) /etc/logrotate.d/*
%dir /etc/tsp-controller
%config(noreplace) /etc/tsp-controller/config
%attr(0755,root,root) /usr/bin/*
/usr/share/man/man*/*
