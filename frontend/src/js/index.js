/*
 *  ZEUS - An Electrifying Build System
 *  Copyright (c) 2017 Philipp Mieden <dreadl0ck@protonmail.ch>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

$(document).ready(function() {

    $('#btn-wiki').click(function() {
        window.open("/wiki", "_blank");
    });

    $('#btn-scripts').click(function() {
        window.open("/scripts", "_blank");
    });

    $('#btn-config').click(function() {
        console.log("toggle config panel");
    });

    $('#btn-quit').click(function() {
        window.location = "/quit";
        setTimeout(function() {
            window.close();
        }, 1000);
    });

	// Create and connect to the server.
    // Optional pass a host string and options.
    var socket = glue();

    socket.onMessage(function(data) {
        console.log("onMessage: " + data);
    });

    socket.on("connected", function() {
        console.log("connected");
    });

    socket.on("connecting", function() {
        console.log("connecting");
    });

    socket.on("disconnected", function() {
        console.log("disconnected");
    });

    socket.on("reconnecting", function() {
        console.log("reconnecting");
    });

    socket.on("error", function(e, msg) {
        console.log("error: " + msg);
    });

    socket.on("connect_timeout", function() {
        console.log("connect_timeout");
    });

    socket.on("timeout", function() {
        console.log("timeout");
    });

    socket.on("discard_send_buffer", function() {
        console.log("some data could not be send and was discarded.");
    });
});

// show spinner
function spinnerON() {
    $('#main-spinner').toggle(true);
}

// hide spinner
function spinnerOFF() {
    $('#main-spinner').toggle(false);
}