/*

#  EQCSS
## version 1.7.0

A JavaScript plugin to read EQCSS syntax to provide:
scoped styles, element queries, container queries,
meta-selectors, eval(), and element-based units.

- github.com/eqcss/eqcss
- elementqueries.com

Authors: Tommy Hodgins, Maxime Euzière, Azareal

License: MIT

*/

// Uses Node, AMD or browser globals to create a module
(function (root, factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD: Register as an anonymous module
    define([], factory);
  } else if (typeof module === 'object' && module.exports) {
    // Node: Does not work with strict CommonJS, but
    // only CommonJS-like environments that support module.exports,
    // like Node
    module.exports = factory();
  } else {
    // Browser globals (root is window)
    root.EQCSS = factory();
  }
}(this, function() {
    var EQCSS = {
      data: []
    }

    /*
     * EQCSS.load()
     * Called automatically on page load.
     * Call it manually after adding EQCSS code in the page.
     * Loads and parses all the EQCSS code.
     */
    EQCSS.load = function() {
      // Retrieve all style blocks
      var styles = document.getElementsByTagName('style');

      for (var i = 0; i < styles.length; i++) {
        // Test if the style is not read yet
        if (styles[i].getAttribute('data-eqcss-read') === null) {

          // Mark the style block as read
          styles[i].setAttribute('data-eqcss-read', 'true');

          EQCSS.process(styles[i].innerHTML);
        }
      }

      // Retrieve all link tags
      var link = document.getElementsByTagName('link');

      for (i = 0; i < link.length; i++) {
        // Test if the link is not read yet, and has rel=stylesheet
        if (link[i].getAttribute('data-eqcss-read') === null && link[i].rel === 'stylesheet') {
          // retrieve the file content with AJAX and process it
          if (link[i].href) {
            (function() {
              var xhr = new XMLHttpRequest;
              xhr.open('GET', link[i].href, true);
              xhr.send(null);
              xhr.onreadystatechange = function() {
                EQCSS.process(xhr.responseText);
              }
            })();
          }
          // Mark the link as read
          link[i].setAttribute('data-eqcss-read', 'true');
        }
      }
    }

    /*
     * EQCSS.parse()
     * Called by load for each script / style / link resource.
     * Generates data for each Element Query found
     */
    EQCSS.parse = function(code) {
      var parsed_queries = new Array();

      // Cleanup
      code = code.replace(/\s+/g, ' '); // reduce spaces and line breaks
      code = code.replace(/\/\*[\w\W]*?\*\//g, ''); // remove comments
      code = code.replace(/@element/g, '\n@element'); // one element query per line
      code = code.replace(/(@element.*?\{([^}]*?\{[^}]*?\}[^}]*?)*\}).*/g, '$1'); // Keep the queries only (discard regular css written around them)

      // Parse

      // For each query
      code.replace(/(@element.*(?!@element))/g, function(string, query) {
        // Create a data entry
        var dataEntry = {};

        // Extract the selector
        query.replace(/(@element)\s*(".*?"|'.*?'|.*?)\s*(and\s*\(|{)/g, function(string, atrule, selector, extra) {
          // Strip outer quotes if present
          selector = selector.replace(/^\s?['](.*)[']/, '$1');
          selector = selector.replace(/^\s?["](.*)["]/, '$1');

          dataEntry.selector = selector;
        })

        // Extract the conditions (measure, value, unit)
        dataEntry.conditions = [];
        query.replace(/and ?\( ?([^:]*) ?: ?([^)]*) ?\)/g, function(string, measure, value) {
          // Separate value and unit if it's possible
          var unit = null;
          unit = value.replace(/^(\d*\.?\d+)(\D+)$/, '$2');

          if (unit === value) {
            unit = null;
          }
          value = value.replace(/^(\d*\.?\d+)\D+$/, '$1');
          dataEntry.conditions.push({measure: measure, value: value, unit: unit});
        });

        // Extract the styles
        query.replace(/{(.*)}/g, function(string, style) {
          dataEntry.style = style;
        });

        parsed_queries.push(dataEntry);
      });

      return parsed_queries;
    }

    /*
     * EQCSS.register()
     * Add a single object, or an array of objects to EQCSS.data
     *
     */
     EQCSS.register = function(queries) {
       if (Object.prototype.toString.call(queries) === '[object Object]') {
         EQCSS.data.push(queries);
         EQCSS.apply();
       }

       if (Object.prototype.toString.call(queries) === '[object Array]') {
        for (var i=0; i<queries.length; i++) {
          EQCSS.data.push(queries[i]);
        }
        EQCSS.apply();
       }
     }

    /*
     * EQCSS.process()
     * Parse and Register queries with `EQCSS.data`
     */

     EQCSS.process = function(code) {
       var queries = EQCSS.parse(code)
       return EQCSS.register(queries)
     }

    /*
     * EQCSS.apply()
     * Called on load, on resize and manually on DOM update
     * Enable the Element Queries in which the conditions are true
     */
    EQCSS.apply = function() {
      var elements;                     // Elements targeted by each query
      var element_guid;                 // GUID for current element
      var css_block;                    // CSS block corresponding to each targeted element
      var element_guid_parent;          // GUID for current element's parent
      var element_guid_prev;            // GUID for current element's previous sibling element
      var element_guid_next;            // GUID for current element's next sibling element
      var css_code;                     // CSS code to write in each CSS block (one per targeted element)
      var element_width, parent_width;  // Computed widths
      var element_height, parent_height;// Computed heights
      var element_line_height;          // Computed line-height
      var test;                         // Query's condition test result
      var computed_style;               // Each targeted element's computed style
      var parent_computed_style;        // Each targeted element parent's computed style

      // Loop on all element queries
      for (var i = 0; i < EQCSS.data.length; i++) {
        // Find all the elements targeted by the query
        elements = document.querySelectorAll(EQCSS.data[i].selector);

        // Loop on all the elements
        for (var j = 0; j < elements.length; j++) {
          // Create a guid for this element
          // Pattern: 'EQCSS_{element-query-index}_{matched-element-index}'
          element_guid = 'data-eqcss-' + i + '-' + j;

          // Add this guid as an attribute to the element
          elements[j].setAttribute(element_guid, '');

          // Create a guid for the parent of this element
          // Pattern: 'EQCSS_{element-query-index}_{matched-element-index}_parent'
          element_guid_parent = 'data-eqcss-' + i + '-' + j + '-parent';

          // Add this guid as an attribute to the element's parent (except if element is the root element)
          if (elements[j] != document.documentElement) {
            elements[j].parentNode.setAttribute(element_guid_parent, '');
          }

          // Get the CSS block associated to this element (or create one in the <HEAD> if it doesn't exist)
          css_block = document.querySelector('#' + element_guid);

          if (!css_block) {
            css_block = document.createElement('style');
            css_block.id = element_guid;
            css_block.setAttribute('data-eqcss-read', 'true');
            document.querySelector('head').appendChild(css_block);
          }
          css_block = document.querySelector('#' + element_guid);

          // Reset the query test's result (first, we assume that the selector is matched)
          test = true;

          // Loop on the conditions
          test_conditions: for (var k = 0; k < EQCSS.data[i].conditions.length; k++) {
            // Reuse element and parent's computed style instead of computing it everywhere
            computed_style = window.getComputedStyle(elements[j], null);

            parent_computed_style = null;

            if (elements[j] != document.documentElement) {
              parent_computed_style = window.getComputedStyle(elements[j].parentNode, null);
            }

            // Do we have to reconvert the size in px at each call?
            // This is true only for vw/vh/vmin/vmax
            var recomputed = false;

            // If the condition's unit is vw, convert current value in vw, in px
            if (EQCSS.data[i].conditions[k].unit === 'vw') {
              recomputed = true;

              var value = parseInt(EQCSS.data[i].conditions[k].value);
              EQCSS.data[i].conditions[k].recomputed_value = value * window.innerWidth / 100;
            }

            // If the condition's unit is vh, convert current value in vh, in px
            else if (EQCSS.data[i].conditions[k].unit === 'vh') {
              recomputed = true;

              var value = parseInt(EQCSS.data[i].conditions[k].value);
              EQCSS.data[i].conditions[k].recomputed_value = value * window.innerHeight / 100;
            }

            // If the condition's unit is vmin, convert current value in vmin, in px
            else if (EQCSS.data[i].conditions[k].unit === 'vmin') {
              recomputed = true;

              var value = parseInt(EQCSS.data[i].conditions[k].value);
              EQCSS.data[i].conditions[k].recomputed_value = value * Math.min(window.innerWidth, window.innerHeight) / 100;
            }

            // If the condition's unit is vmax, convert current value in vmax, in px
            else if (EQCSS.data[i].conditions[k].unit === 'vmax') {
              recomputed = true;

              var value = parseInt(EQCSS.data[i].conditions[k].value);
              EQCSS.data[i].conditions[k].recomputed_value = value * Math.max(window.innerWidth, window.innerHeight) / 100;
            }

            // If the condition's unit is set and is not px or %, convert it into pixels
            else if (EQCSS.data[i].conditions[k].unit != null && EQCSS.data[i].conditions[k].unit != 'px' && EQCSS.data[i].conditions[k].unit != '%') {
              // Create a hidden DIV, sibling of the current element (or its child, if the element is <html>)
              // Set the given measure and unit to the DIV's width
              // Measure the DIV's width in px
              // Remove the DIV
              var div = document.createElement('div');

              div.style.visibility = 'hidden';
              div.style.border = '1px solid red';
              div.style.width = EQCSS.data[i].conditions[k].value + EQCSS.data[i].conditions[k].unit;

              var position = elements[j];
              if (elements[j] != document.documentElement) {
                position = elements[j].parentNode;
              }

              position.appendChild(div);
              EQCSS.data[i].conditions[k].value = parseInt(window.getComputedStyle(div, null).getPropertyValue('width'));
              EQCSS.data[i].conditions[k].unit = 'px';
              position.removeChild(div);
            }

            // Store the good value in final_value depending if the size is recomputed or not
            var final_value = recomputed ? EQCSS.data[i].conditions[k].recomputed_value : parseInt(EQCSS.data[i].conditions[k].value);

            // Check each condition for this query and this element
            // If at least one condition is false, the element selector is not matched
            switch (EQCSS.data[i].conditions[k].measure) {
             case 'min-width':
                // Min-width in px
                if (recomputed === true || EQCSS.data[i].conditions[k].unit === 'px') {
                  element_width = parseInt(computed_style.getPropertyValue('width'));
                  if (!(element_width >= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }

                // Min-width in %
                if (EQCSS.data[i].conditions[k].unit === '%') {
                  element_width = parseInt(computed_style.getPropertyValue('width'));
                  parent_width = parseInt(parent_computed_style.getPropertyValue('width'));
                  if (!(parent_width / element_width <= 100 / final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }

              break;

              case 'max-width':
                // Max-width in px
                if (recomputed === true || EQCSS.data[i].conditions[k].unit === 'px') {
                  element_width = parseInt(computed_style.getPropertyValue('width'));
                  if (!(element_width <= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }

                // Max-width in %
                if (EQCSS.data[i].conditions[k].unit === '%') {
                  element_width = parseInt(computed_style.getPropertyValue('width'));
                  parent_width = parseInt(parent_computed_style.getPropertyValue('width'));
                  if (!(parent_width / element_width >= 100 / final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }
              break;

              case 'min-height':
                // Min-height in px
                if (recomputed === true || EQCSS.data[i].conditions[k].unit === 'px') {
                  element_height = parseInt(computed_style.getPropertyValue('height'));
                  if (!(element_height >= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }

                // Min-height in %
                if (EQCSS.data[i].conditions[k].unit === '%') {
                  element_height = parseInt(computed_style.getPropertyValue('height'));
                  parent_height = parseInt(parent_computed_style.getPropertyValue('height'));
                  if (!(parent_height / element_height <= 100 / final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }
              break;

              case 'max-height':
                // Max-height in px
                if (recomputed === true || EQCSS.data[i].conditions[k].unit === 'px') {
                  element_height = parseInt(computed_style.getPropertyValue('height'));
                  if (!(element_height <= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }

                // Max-height in %
                if (EQCSS.data[i].conditions[k].unit === '%') {
                  element_height = parseInt(computed_style.getPropertyValue('height'));
                  parent_height = parseInt(parent_computed_style.getPropertyValue('height'));
                  if (!(parent_height / element_height >= 100 / final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }
              break;
              
              // Min-characters
              case 'min-characters':
                // form inputs
                if (elements[j].value) {
                  if (!(elements[j].value.length >= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }
                // blocks
                else {
                  if (!(elements[j].textContent.length >= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }
              break;

              // Max-characters
              case 'max-characters':
                // form inputs
                if (elements[j].value) {
                  if (!(elements[j].value.length <= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }
                // blocks
                else {
                  if (!(elements[j].textContent.length <= final_value)) {
                    test = false;
                    break test_conditions;
                  }
                }
              break;

              // Min-children
              case 'min-children':
                if (!(elements[j].children.length >= final_value)) {
                  test = false;
                  break test_conditions;
                }
              break;

              // Max-children
              case 'max-children':
                if (!(elements[j].children.length <= final_value)) {
                  test = false;
                  break test_conditions;
                }
              break;

              // Min-lines
              case 'min-lines':
                element_height =
                  parseInt(computed_style.getPropertyValue('height'))
                  - parseInt(computed_style.getPropertyValue('border-top-width'))
                  - parseInt(computed_style.getPropertyValue('border-bottom-width'))
                  - parseInt(computed_style.getPropertyValue('padding-top'))
                  - parseInt(computed_style.getPropertyValue('padding-bottom'));

                element_line_height = computed_style.getPropertyValue('line-height');
                if (element_line_height === 'normal') {
                  var element_font_size = parseInt(computed_style.getPropertyValue('font-size'));
                  element_line_height = element_font_size * 1.125;
                } else {
                  element_line_height = parseInt(element_line_height);
                }

                if (!(element_height / element_line_height >= final_value)) {
                  test = false;
                  break test_conditions;
                }
              break;

              // Max-lines
              case 'max-lines':
                element_height =
                  parseInt(computed_style.getPropertyValue('height'))
                  - parseInt(computed_style.getPropertyValue('border-top-width'))
                  - parseInt(computed_style.getPropertyValue('border-bottom-width'))
                  - parseInt(computed_style.getPropertyValue('padding-top'))
                  - parseInt(computed_style.getPropertyValue('padding-bottom'));

                element_line_height = computed_style.getPropertyValue('line-height');
                if (element_line_height === 'normal') {
                  var element_font_size = parseInt(computed_style.getPropertyValue('font-size'));
                  element_line_height = element_font_size * 1.125;
                } else {
                  element_line_height = parseInt(element_line_height);
                }

                if (!(element_height / element_line_height + 1 <= final_value)) {
                  test = false;
                  break test_conditions;
                }
              break;
            }
          }

          // Update CSS block:
          // If all conditions are met: copy the CSS code from the query to the corresponding CSS block
          if (test === true) {
            // Get the CSS code to apply to the element
            css_code = EQCSS.data[i].style;

            // Replace eval('xyz') with the result of try{with(element){eval(xyz)}} in JS
            css_code = css_code.replace(
              /eval\( *((".*?")|('.*?')) *\)/g,
              function(string, match) {
                return EQCSS.tryWithEval(elements[j], match);
              }
            );

            // Replace '$this' or 'eq_this' with '[element_guid]'
            css_code = css_code.replace(/(\$|eq_)this/gi, '[' + element_guid + ']');

            // Replace '$parent' or 'eq_parent' with '[element_guid_parent]'
            css_code = css_code.replace(/(\$|eq_)parent/gi, '[' + element_guid_parent + ']');
            
            if(css_block.innerHTML != css_code){
              css_block.innerHTML = css_code;
            }
          }

          // If condition is not met: empty the CSS block
          else if(css_block.innerHTML != '') {
            css_block.innerHTML = '';
          }
        }
      }
    }

    /*
     * Eval('') and $it
     * (…yes with() was necessary, and eval() too!)
     */
    EQCSS.tryWithEval = function(element, string) {
      var $it = element;
      var ret = '';
      
      try {
        with ($it) { ret = eval(string.slice(1, -1)) }
      }
      catch(e) {
        ret = '';
      }
      return ret;
    }

    /*
     * EQCSS.reset
     * Deletes parsed queries removes EQCSS-generated tags and attributes
     * To reload EQCSS again after running EQCSS.reset() use EQCSS.load()
     */
    EQCSS.reset = function() {
      // Reset EQCSS.data, removing previously parsed queries
      EQCSS.data = [];
      
      // Remove EQCSS-generated style tags from head
      var style_tag = document.querySelectorAll('head style[id^="data-eqcss-"]');
      for (var i = 0; i < style_tag.length; i++) {
        style_tag[i].parentNode.removeChild(style_tag[i]);
      }

      // Remove EQCSS-generated attributes from all tags
      var tag = document.querySelectorAll('*');

      // For each tag in the document
      for (var j = 0; j < tag.length; j++) {
        // Loop through all attributes
        for (var k = 0; k < tag[j].attributes.length; k++) {
          // If an attribute begins with 'data-eqcss-'
          if (tag[j].attributes[k].name.indexOf('data-eqcss-') === 0) {
            // Remove the attribute from the tag
            tag[j].removeAttribute(tag[j].attributes[k].name)
          }
        }
      }
    }

    /*
     * 'DOM Ready' cross-browser polyfill / Diego Perini / MIT license
     * Forked from: https://github.com/dperini/ContentLoaded/blob/master/src/contentloaded.js
     */
    EQCSS.domReady = function(fn) {
      var done = false;
      var top = true;
      var doc = window.document;
      var root = doc.documentElement;
      var modern = !~navigator.userAgent.indexOf('MSIE 8');
      var add = modern ? 'addEventListener' : 'attachEvent';
      var rem = modern ? 'removeEventListener' : 'detachEvent';
      var pre = modern ? '' : 'on';
      var init = function(e) {
        if (e.type === 'readystatechange' && doc.readyState !== 'complete') return;
        (e.type === 'load' ? window : doc)[rem](pre + e.type, init, false);
        if (!done && (done = true)) fn.call(window, e.type || e);
      },
      poll = function() {
        try {
          root.doScroll('left');
        }
        catch(e) {
          setTimeout(poll, 50);
          return;
        }
        init('poll');
      };

      if (doc.readyState === 'complete') {
        fn.call(window, 'lazy');
        return;
      }
      
      if (!modern && root.doScroll) {
          try {
            top = !window.frameElement;
          }
          catch(e) {}
          if (top) poll();
      }
      doc[add](pre + 'DOMContentLoaded', init, false);
      doc[add](pre + 'readystatechange', init, false);
      window[add](pre + 'load', init, false);
    }

    /*
     * EQCSS.throttle
     * Ensures EQCSS.apply() is not called more than once every (EQCSS_timeout)ms
     */
    var EQCSS_throttle_available = true;
    var EQCSS_throttle_queued = false;
    var EQCSS_mouse_down = false;
    var EQCSS_timeout = 200;

    EQCSS.throttle = function() {
     /* if (EQCSS_throttle_available) {*/
        EQCSS.apply();
        /*EQCSS_throttle_available = false;

        setTimeout(function() {
          EQCSS_throttle_available = true;
          if (EQCSS_throttle_queued) {
            EQCSS_throttle_queued = false;
            EQCSS.apply();
          }
        }, EQCSS_timeout);
      } else {
        EQCSS_throttle_queued = true;
      }*/
    }

    // Call load (and apply, indirectly) on page load
    EQCSS.domReady(function() {
      EQCSS.load();
      EQCSS.throttle();
    });

    // On resize, click, call EQCSS.throttle.
    window.addEventListener('resize', EQCSS.throttle);
    window.addEventListener('click', EQCSS.throttle);

    // Debug: here's a shortcut for console.log
    function l(a) { console.log(a) }

    return EQCSS;
}));