# vim: syntax=spec

%if %{with static}
%global build_static true
%else
%global build_static false
%endif

Name:       faq
Version:    0.0.4
Release:    1%{?dist}
Summary:    Command-line JSON/YAML/XML/TOML/BSON processor
License:    Apache 2.0
URL:        https://github.com/jzelinskie/faq
Source0:    https://github.com/chancez/faq/archive/%{branch}/%{name}-%{version}.tar.gz

BuildRequires: gcc
BuildRequires: make
BuildRequires: golang >= 1.11
BuildRequires: oniguruma-devel
BuildRequires: jq-devel >= 1.6
%if %{without static}
Requires: jq >= 1.6
%endif

ExclusiveArch: x86_64

%description
Format Agnostic jQ

%prep
%autosetup

%build

%make_build SKIP_VALIDATE=true FAQ_LINK_STATIC=%{build_static} GO_LD_FLAGS='-compressdwarf=false -linkmode external -extldflags "-v"'

%install
rm -rf $RPM_BUILD_ROOT
%make_install prefix=%{_prefix}

%files
%license LICENSE
%{_bindir}/%{name}

%changelog
* Tue Jan 29 2019 Chance Zibolski <czibolsk@redhat.com> 0.0.4-1
- First faq package
