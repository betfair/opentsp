# Support short-hand rpmbuild invocation.
%define _sourcedir .
%define _builddir .
%define _srcrpmdir .
%define _rpmdir .

# Disable rpm magic that creates junk files.
%define __debug_install_post %{nil}
%define __os_install_post %{nil}

Name: collect-statse
Version: %{version}
Release: 1
Summary: collect Statse metrics
Group: Systems
License: Proprietary
Packager: Jacek Masiulaniec <jacek.masiulaniec@gmail.com>
BuildRoot: /var/tmp/%{name}-buildroot
Conflicts: kernel = 1:2.6.18-194.32.1.el5
Conflicts: kernel = 1:2.6.18-238
Conflicts: kernel = 1:2.6.18-238.5.1.el5
Conflicts: kernel-xen = 2.6.18-238.12.1.el5
Conflicts: kernel-xen = 1:2.6.18-238.12.1.el5
Conflicts: kernel-xen = 1:2.6.18-274.7.1.el5

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

%description

%files
%defattr(644,root,root,755)
%attr(0755,root,root) /usr/bin/*
/usr/share/man/man*/*
