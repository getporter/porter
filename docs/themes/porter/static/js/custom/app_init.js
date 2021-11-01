$(document).ready(function() {
  $(document).foundation();

  // custom theme js for sidebar links
  var allClosed;

  // close all accordions, besides that of the page that is currently active
  var curPage = window.location.pathname+window.location.search+window.location.hash

  // Find nav links even when they don't end in /
  var curPage2 = curPage.slice(0, curPage.length-1)
  var activeLink = $(".sidebar-nav a[href='" + curPage + "'], .sidebar-nav a[href='" + curPage2 + "']");
  activeLink.addClass('active');

  // Try to open the parent menus
  var parentMenu = activeLink.closest('li.toctree-l2');
  if(parentMenu) {
    var parentLink = parentMenu.children('a');
    parentLink.addClass('active').attr({state: "open"});
  }

  var parentMenu = activeLink.closest('li.toctree-l1');
  var parentLink = parentMenu.children('a');
  parentLink.addClass('active').attr({state: "open"});

  if (allClosed === true) { }

  // if menu is closed when clicked, expand it
  $('.toctree-l1 > a').click(function() {
    //Make the titles of open accordions dead links
    if ($(this).attr('state') == 'open') {return false;}

    //Clicking on a title of a closed accordion
    if($(this).attr('state') != 'open' && $(this).siblings().size() > 0) {
      $('.toctree-l1 > ul, .toctree-l2 > ul').hide();
      $('.toctree-l1 > a, .toctree-l2 > a').attr('state', '');
      $(this).attr('state', 'open');
      let nestedTrees = this.nextElementSibling.querySelectorAll('ul');
      for (let i = 0; i < nestedTrees.length; i++) {
        let tree = nestedTrees[0]
        tree.style.display = "block"
      }
      $(this).next().slideDown(function(){});
      return false;
    }
  });
}); // document ready


// add permalinks to titles
$(function() {
  return $("h1, h2, h3, h4, h5, h6").each(function(i, el) {
    var $el, icon, id;
    $el = $(el);
    id = $el.attr('id');
    icon = '<i class="fa fa-link"></i>';
    if (id) {
      return $el.prepend($("<a />").addClass("header-link").attr("href", "#" + id).html(icon));
    }
  });
});
