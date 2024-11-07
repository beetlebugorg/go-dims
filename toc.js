// Populate the sidebar
//
// This is a script, and not included directly in the page, to control the total size of the book.
// The TOC contains an entry for each page, so if each page includes a copy of the TOC,
// the total size of the page becomes O(n**2).
var sidebarScrollbox = document.querySelector("#sidebar .sidebar-scrollbox");
sidebarScrollbox.innerHTML = '<ol class="chapter"><li class="chapter-item expanded affix "><a href="index.html">Introduction</a></li><li class="chapter-item expanded affix "><li class="part-title">User Guide</li><li class="chapter-item expanded "><a href="guide/installation.html"><strong aria-hidden="true">1.</strong> Getting Started</a></li><li class="chapter-item expanded "><a href="guide/running.html"><strong aria-hidden="true">2.</strong> Running</a></li><li><ol class="section"><li class="chapter-item expanded "><a href="running/aws.html"><strong aria-hidden="true">2.1.</strong> AWS</a></li><li class="chapter-item expanded "><a href="running/local.html"><strong aria-hidden="true">2.2.</strong> Local</a></li><li class="chapter-item expanded "><a href="running/resources.html"><strong aria-hidden="true">2.3.</strong> Resources</a></li></ol></li><li class="chapter-item expanded "><a href="configuration/index.html"><strong aria-hidden="true">3.</strong> Configuration</a></li><li><ol class="section"><li class="chapter-item expanded "><a href="configuration/signing.html"><strong aria-hidden="true">3.1.</strong> Signing Requests</a></li><li class="chapter-item expanded "><a href="configuration/cache-control.html"><strong aria-hidden="true">3.2.</strong> Cache Control Headers</a></li><li class="chapter-item expanded "><a href="configuration/other.html"><strong aria-hidden="true">3.3.</strong> Other Settings</a></li><li class="chapter-item expanded "><a href="configuration/mod-dims.html"><strong aria-hidden="true">3.4.</strong> mod-dims</a></li></ol></li><li class="chapter-item expanded "><div><strong aria-hidden="true">4.</strong> Administration</div></li><li><ol class="section"><li class="chapter-item expanded "><div><strong aria-hidden="true">4.1.</strong> Security</div></li><li><ol class="section"><li class="chapter-item expanded "><div><strong aria-hidden="true">4.1.1.</strong> Vulnerabilities</div></li><li class="chapter-item expanded "><div><strong aria-hidden="true">4.1.2.</strong> Upgrades</div></li></ol></li><li class="chapter-item expanded "><div><strong aria-hidden="true">4.2.</strong> Logging</div></li><li class="chapter-item expanded "><div><strong aria-hidden="true">4.3.</strong> Metrics</div></li><li class="chapter-item expanded "><div><strong aria-hidden="true">4.4.</strong> Monitoring</div></li></ol></li><li class="chapter-item expanded "><li class="part-title">API Reference</li><li class="chapter-item expanded "><div><strong aria-hidden="true">5.</strong> Endpoints</div></li><li><ol class="section"><li class="chapter-item expanded "><div><strong aria-hidden="true">5.1.</strong> /v4</div></li><li><ol class="section"><li class="chapter-item expanded "><a href="endpoints/dims4.html"><strong aria-hidden="true">5.1.1.</strong> /v4/dims</a></li></ol></li><li class="chapter-item expanded "><div><strong aria-hidden="true">5.2.</strong> /v5</div></li><li><ol class="section"><li class="chapter-item expanded "><a href="endpoints/dims5.html"><strong aria-hidden="true">5.2.1.</strong> /v5/dims</a></li></ol></li><li class="chapter-item expanded "><a href="endpoints/status.html"><strong aria-hidden="true">5.3.</strong> /status</a></li><li class="chapter-item expanded "><a href="endpoints/health.html"><strong aria-hidden="true">5.4.</strong> /health</a></li></ol></li><li class="chapter-item expanded "><div><strong aria-hidden="true">6.</strong> Image Management Commands</div></li><li><ol class="section"><li class="chapter-item expanded "><a href="operations/resize.html"><strong aria-hidden="true">6.1.</strong> Resize</a></li><li class="chapter-item expanded "><a href="operations/crop.html"><strong aria-hidden="true">6.2.</strong> Crop</a></li><li class="chapter-item expanded "><a href="operations/thumbnail.html"><strong aria-hidden="true">6.3.</strong> Thumbnail</a></li><li class="chapter-item expanded "><a href="operations/quality.html"><strong aria-hidden="true">6.4.</strong> Quality</a></li><li class="chapter-item expanded "><a href="operations/format.html"><strong aria-hidden="true">6.5.</strong> Format</a></li><li class="chapter-item expanded "><a href="operations/strip.html"><strong aria-hidden="true">6.6.</strong> Strip</a></li><li class="chapter-item expanded "><a href="operations/sharpen.html"><strong aria-hidden="true">6.7.</strong> Sharpen</a></li><li class="chapter-item expanded "><a href="operations/brightness.html"><strong aria-hidden="true">6.8.</strong> Brightness</a></li><li class="chapter-item expanded "><a href="operations/flipflop.html"><strong aria-hidden="true">6.9.</strong> Flip Flop</a></li><li class="chapter-item expanded "><a href="operations/sepia.html"><strong aria-hidden="true">6.10.</strong> Sepia</a></li><li class="chapter-item expanded "><a href="operations/grayscale.html"><strong aria-hidden="true">6.11.</strong> Grayscale</a></li><li class="chapter-item expanded "><a href="operations/invert.html"><strong aria-hidden="true">6.12.</strong> Invert</a></li><li class="chapter-item expanded "><a href="operations/rotate.html"><strong aria-hidden="true">6.13.</strong> Rotate</a></li><li class="chapter-item expanded "><a href="operations/legacy_thumbnail.html"><strong aria-hidden="true">6.14.</strong> Legacy Thumbnail</a></li><li class="chapter-item expanded "><a href="operations/gravity.html"><strong aria-hidden="true">6.15.</strong> Gravity</a></li></ol></li></ol>';
(function() {
    let current_page = document.location.href.toString();
    if (current_page.endsWith("/")) {
        current_page += "index.html";
    }
    var links = sidebarScrollbox.querySelectorAll("a");
    var l = links.length;
    for (var i = 0; i < l; ++i) {
        var link = links[i];
        var href = link.getAttribute("href");
        if (href && !href.startsWith("#") && !/^(?:[a-z+]+:)?\/\//.test(href)) {
            link.href = path_to_root + href;
        }
        // The "index" page is supposed to alias the first chapter in the book.
        if (link.href === current_page || (i === 0 && path_to_root === "" && current_page.endsWith("/index.html"))) {
            link.classList.add("active");
            var parent = link.parentElement;
            while (parent) {
                if (parent.tagName === "LI" && parent.previousElementSibling) {
                    if (parent.previousElementSibling.classList.contains("chapter-item")) {
                        parent.previousElementSibling.classList.add("expanded");
                    }
                }
                parent = parent.parentElement;
            }
        }
    }
})();

// Track and set sidebar scroll position
sidebarScrollbox.addEventListener('click', function(e) {
    if (e.target.tagName === 'A') {
        sessionStorage.setItem('sidebar-scroll', sidebarScrollbox.scrollTop);
    }
}, { passive: true });
var sidebarScrollTop = sessionStorage.getItem('sidebar-scroll');
sessionStorage.removeItem('sidebar-scroll');
if (sidebarScrollTop) {
    // preserve sidebar scroll position when navigating via links within sidebar
    sidebarScrollbox.scrollTop = sidebarScrollTop;
} else {
    // scroll sidebar to current active section when navigating via "next/previous chapter" buttons
    var activeSection = document.querySelector('#sidebar .active');
    if (activeSection) {
        activeSection.scrollIntoView({ block: 'center' });
    }
}
