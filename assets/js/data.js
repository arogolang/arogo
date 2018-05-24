'use strict';


function loadIndexData() {
    $.ajax({
        type: "GET",
        dataType: "json",
        url: "api.php?q=index",
        success: function(data) {
            if (data) {
                var html = "";
                for(var item in data.shares) {
                    html = "<tr><td>"+item.shares+"</td><td>"+item.percent+"</td><td>"+item.bestdl+"</td><td>"+item.id+"</td></tr>";
                }
                $("#table-miner").html(html);

                html = "";
                for(var item in data.historics) {
                    html = "<tr><td>"+item.historic+"</td><td>"+item.percent+"</td><td>"+item.hashrate+"</td><td>"+item.pending+"</td><td>"+item.totalpaid+"</td><td>"+item.id+"</td></tr>";
                }

                $("#table-historic").html(html);

                $("#cur-block-height").html(data.height);
                $("#last-won-block").html(data.lastwon);
                $("#cur-miners").html(data.miners);
                $("#cur-difficulty").html(data.difficulty);

                $("#cur-total-paid").html(data.totalpaid + "M");
                $("#cur-total-hr").html(data.totalhr);
                $("#cur-avg-hr").html(data.avghr);

                $("#cur-total-shares").html("Total Current Shares: " + data.totalshares);
                $("#cur-total-historic").html("Total Historic Shares: " + data.totalhistoric);
                
            }
        }
    });    
}

function loadBlocksData() {
    $.ajax({
        type: "GET",
        dataType: "json",
        url: "api.php?q=blocks",
        success: function(data) {
            if (data) {
                var html = "";
                for(var item in data) {
                    html = "<tr><td>"+item.id+"</td><td>"+item.height+"</td><td>"+item.reward+"</td><td>"+item.miner+"</td></tr>";
                }

                $("#blocks-table").html(html);
            }
        }
    });    
}

function loadPaymentsData() {
    $.ajax({
        type: "GET",
        dataType: "json",
        url: "api.php?q=payments",
        success: function(data) {
            if (data) {
                var html = "";
                for(var item in data) {
                    html = "<tr><td>"+item.id+"</td><td>"+item.address+"</td><td>"+item.val+"</td><td>"+item.txn+"</td></tr>";
                }

                $("#payments-table").html(html);
            }
        }
    });
}

function setMenuActive(){
    if ( q == undefined) {
        q = ""
    }
    

    $("#menuq"+q).addClass("active");

    if (q == "payments") {
        loadPaymentsData();
        $('.datatable').dataTable({"order": [[ 0, "desc" ]], "iDisplayLength": 15});
    }
    else if(q == "blocks") {
        loadBlocksData();
        $('.datatable').dataTable({"order": [[ 1, "desc" ]], "iDisplayLength": 15});
    }
    else if ( q=="") {
        loadIndexData();        
        $('.datatable').dataTable({"iDisplayLength": 15});
    }
}

$(document).ready(function () {
    setMenuActive();
});